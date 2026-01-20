package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connectionString := "amqp://guest:guest@localhost:5672/"
	fmt.Println("Starting Peril client...")

	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	fmt.Println("Connection to broker successful.")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	amqpChannel, _, err := pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		"pause."+username,
		routing.PauseKey,
		pubsub.TRANSIENT,
	)

	gameState := gamelogic.NewGameState(username)
	pubsub.SubscribeJSON(connection,
		routing.ExchangePerilDirect,
		"pause."+username,
		routing.PauseKey,
		pubsub.TRANSIENT,
		handlerPause(gameState),
	)
	pubsub.SubscribeJSON(connection,
		routing.ExchangePerilTopic,
		"army_moves."+username,
		"army_moves.*",
		pubsub.TRANSIENT,
		handlerMove(gameState, amqpChannel),
	)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

mainloop:
	for {
		inputWords := gamelogic.GetInput()
		switch inputWords[0] {
		case "spawn":
			fmt.Println("Spawning troops...")
			err := gameState.CommandSpawn(inputWords)
			if err != nil {
				fmt.Printf("Error in the spawn command: %v\n", err)
			}
		case "move":
			fmt.Println("Moving troops...")
			move, err := gameState.CommandMove(inputWords)
			if err != nil {
				fmt.Printf("Error in the move command: %v\n", err)
			} else {
				pubsub.PublishJSON(
					amqpChannel,
					routing.ExchangePerilTopic,
					"army_moves."+username,
					move,
				)
				fmt.Println("Move successful!")
			}
		case "status":
			fmt.Println("Printing player status...")
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break mainloop
		default:
			fmt.Println("Command not understood.")
		}
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.ACK
	}
}

func handlerMove(gs *gamelogic.GameState, amqpChan *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.ACK
		case gamelogic.MoveOutcomeMakeWar:
			pubsub.PublishJSON(
				amqpChan,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			return pubsub.NACKREQUEUE
		default:
			return pubsub.NACKDISCARD
		}
	}
}
