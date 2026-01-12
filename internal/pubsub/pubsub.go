package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	DURABLE   SimpleQueueType = "durable"
	TRANSIENT SimpleQueueType = "transient"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonData, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal val into a JSON: %v", err)
	}
	message := amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonData,
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, message)

	if err != nil {
		return fmt.Errorf("failed to publish the message: %v")
	}
	return nil
}
