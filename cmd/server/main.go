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
	const connString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("❌ Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("✅ Successfully connected to RabbitMQ")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("❌ Failed to open a channel: %v", err)
	}

	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*", // "game_logs.*"
		pubsub.QueueDurable,
		func(log routing.GameLog) pubsub.AckType {
			defer fmt.Print("> ")
			if err := gamelogic.WriteLog(log); err != nil {
				fmt.Printf("❌ Failed to write log: %v\n", err)
				return pubsub.NackDiscard
			}
			fmt.Printf("📜 Log written for %s: %s\n", log.Username, log.Message)
			return pubsub.Ack
		},
	)
	if err != nil {
		log.Fatalf("❌ Failed to bind to game_logs.*: %v", err)
	}
	fmt.Println("📥 Waiting for pause/resume commands...")
	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			fmt.Println("⏸️ Sending pause message...")
			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			)
			if err != nil {
				log.Fatalf("❌ Failed to publish message: %v", err)
			}
		case "resume":
			fmt.Println("▶️ Sending resume message...")
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			)
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
