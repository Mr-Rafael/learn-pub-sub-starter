package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerGameLog(log routing.GameLog) pubsub.Acktype {
	defer fmt.Print("> ")
	err := gamelogic.WriteLog(log)
	if err != nil {
		return pubsub.NackDiscard
	}
	return pubsub.Ack
}
