package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connString := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("❌ Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Failed to open a channel: %v", err)
	}
	defer ch.Close()

	fmt.Println("✅ Successfully connected to RabbitMQ")

	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.PauseKey+"*",
		pubsub.QueueDurable,
	)
	if err != nil {
		log.Fatalf("❌ Failed to declare and bind queue: %v", err)
	}
	fmt.Printf("📥 Waiting for messages on queue %q\n", queue.Name)
	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			fmt.Println("⏸️ Sending pause message...")
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			)
			if err != nil {
				log.Fatalf("❌ Failed to publish message: %v", err)
			}
		case "resume":
			fmt.Println("▶️ Sending resume message...")
			err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			if err != nil {
				log.Fatalf("❌ Failed to publish message: %v", err)
			}
		case "quit":
			fmt.Println("👋 Exiting...")
			return
		default:
			fmt.Printf("❓ Unknown command: %q\n", words[0])
		}

	}
}
