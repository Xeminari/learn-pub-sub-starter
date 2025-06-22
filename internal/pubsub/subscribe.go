package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	msgs, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for msg := range msgs {
			var payload T
			err := json.Unmarshal(msg.Body, &payload)
			if err != nil {
				log.Printf("❌ Failed to unmarshal message: %v", err)
				_ = msg.Nack(false, false)
				continue
			}
			switch handler(payload) {
			case Ack:
				_ = msg.Ack(false)
				log.Printf("✅ Acknowledged!")
			case NackRequeue:
				_ = msg.Nack(false, true)
				log.Printf("✅ Nack and Reque!")
			case NackDiscard:
				_ = msg.Nack(false, false)
				log.Printf("✅ Nack and Discard!")
			}
		}
	}()
	return nil
}
