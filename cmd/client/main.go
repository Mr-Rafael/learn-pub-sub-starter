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

	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		"pause."+username,
		routing.PauseKey,
		pubsub.TRANSIENT,
	)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	sig := <-signalChannel

	fmt.Printf("%v signal received. Shutting down...", sig)
	connection.Close()
}
