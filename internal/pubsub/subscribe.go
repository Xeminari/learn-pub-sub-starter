package pubsub

import (
	"bytes"
	"encoding/gob"
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
	unmarshaller := func(data []byte) (T, error) {
		var v T
		err := json.Unmarshal(data, &v)
		return v, err
	}
	return subscribe(conn, exchange, queueName, key, queueType, handler, unmarshaller)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	unmarshaller := func(data []byte) (T, error) {
		var v T
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		err := dec.Decode(&v)
		return v, err
	}
	return subscribe(conn, exchange, queueName, key, queueType, handler, unmarshaller)
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}
	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}
	msgs, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			payload, err := unmarshaller(msg.Body)
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
				log.Printf("🔁 Nack and Requeue!")
			case NackDiscard:
				_ = msg.Nack(false, false)
				log.Printf("🗑️ Nack and Discard!")
			}
		}
	}()
	return nil
}
