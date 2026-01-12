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
	fmt.Println("Starting Peril server...")

	gamelogic.PrintServerHelp()

	connection, err := amqp.Dial(connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	fmt.Println("Connection to broker successful.")

	amqpChannel, err := connection.Channel()
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.DURABLE,
	)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

mainloop:
	for {
		inputWords := gamelogic.GetInput()
		switch inputWords[0] {
		case "pause":
			fmt.Println("Sending a pause message.")
			pubsub.PublishJSON(amqpChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
		case "resume":
			fmt.Println("Sending a resume message.")
			pubsub.PublishJSON(amqpChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
		case "quit":
			fmt.Println("Exiting...")
			break mainloop
		default:
			fmt.Println("Command not understood.")
		}
	}
}
