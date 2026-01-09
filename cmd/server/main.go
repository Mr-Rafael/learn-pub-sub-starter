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

	pubsub.PublishJSON(amqpChannel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	sig := <-signalChannel

	fmt.Printf("%v signal received. Shutting down...", sig)
	connection.Close()
}
