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

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	amqpChannel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to create AMQP channel: %v", err)
	}

	newQueue, err := amqpChannel.QueueDeclare(queueName, queueType == DURABLE, queueType == TRANSIENT, queueType == TRANSIENT, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to create a new queue: %v", err)
	}

	err = amqpChannel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to bind the queue '%v' to exchange '%v': %v", queueName, exchange, err)
	}

	return amqpChannel, newQueue, nil
}
