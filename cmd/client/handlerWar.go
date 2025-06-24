package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func HandlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(msg gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(msg)
		var message string
		var ack pubsub.AckType
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
			message = fmt.Sprintf("%s won a war against %s", winner, loser)
			ack = pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			message = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			ack = pubsub.Ack
		default:
			fmt.Printf("❌ Unknown war outcome: %v\n", outcome)
			return pubsub.NackDiscard
		}
		fmt.Println(message)

		log := routing.GameLog{
			CurrentTime: time.Now(),
			Message:     message,
			Username:    msg.Attacker.Username,
		}
		err := pubsub.PublishGob(
			ch,
			routing.ExchangePerilTopic,
			routing.GameLogSlug+"."+msg.Attacker.Username,
			log,
		)
		if err != nil {
			fmt.Printf("❌ Failed to publish game log: %v\n", err)
			return pubsub.NackRequeue
		}
		return ack
	}
}
