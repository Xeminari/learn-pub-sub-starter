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
		log.Fatalf("❌ Failed to create AMQP channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("❌ Failed to connect to user: %v", err)
	}

	gs := gamelogic.NewGameState(username)

	pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.QueueTransient,
		HandlerMove(gs),
	)

	pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.QueueTransient,
		HandlerPause(gs),
	)

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err := gs.CommandSpawn(words)
			if err != nil {
				fmt.Printf("❌ Failed to spawn unit: %v\n", err)
				continue
			}
		case "move":
			move, err := gs.CommandMove(words)
			if err != nil {
				fmt.Printf("❌ Failed to move: %v\n", err)
				continue
			}
			fmt.Println("✅ You successfully moved!")

			err = pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+username,
				move,
			)
			if err != nil {
				fmt.Printf("❌ Failed to publish move: %v\n", err)
			} else {
				fmt.Println("📤 Move published successfully.")
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming is not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("❓ Unknown command: %q\n", words[0])
		}
	}

}
