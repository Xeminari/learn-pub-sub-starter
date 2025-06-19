package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T),
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
				continue
			}
			handler(payload)
			if err := msg.Ack(false); err != nil {
				log.Printf("❌ Failed to ack message: %v", err)
			}
		}
	}()
	return nil
}
