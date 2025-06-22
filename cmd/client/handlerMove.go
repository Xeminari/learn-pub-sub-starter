package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func HandlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(am gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(am)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			player := gs.GetPlayerSnap()
			war := gamelogic.RecognitionOfWar{
				Attacker: am.Player,
				Defender: player,
			}
			routingKey := routing.ArmyMovesPrefix + "." + am.Player.Username
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routingKey,
				war,
			)
			if err != nil {
				fmt.Printf("❌ Failed to publish war recognition: %v\n", err)
				return pubsub.NackRequeue // Retry on failure
			}
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}
