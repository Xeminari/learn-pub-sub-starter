package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

type simpleQueueType int

const (
	QueueDurable simpleQueueType = iota
	QueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType simpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {

	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not create channel: %w", err)
	}

	queue, err := ch.QueueDeclare(
		queueName,
		queueType == QueueDurable, //durable
		queueType != QueueDurable, //autodelete
		queueType != QueueDurable, //exclusive
		false,
		nil)
	if err != nil {
		ch.Close()
		return nil, amqp.Queue{}, fmt.Errorf("could not declare queue: %w", err)
	}

	err = ch.QueueBind(
		queueName,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		return nil, amqp.Queue{}, fmt.Errorf("could not bind queue: %w", err)
	}
	return ch, queue, nil
}
