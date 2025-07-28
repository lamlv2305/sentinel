package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"

	"github.com/lamlv2305/sentinel/agent"
	"github.com/lamlv2305/sentinel/persister"
	"github.com/lamlv2305/sentinel/types"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
}

func main() {
	persister, err := persister.NewSQLitePersister[types.Resource]("/tmp/agent.db")
	if err != nil {
		panic(err)
	}

	subscriber := make(chan types.Resource, 1) // Buffered channel for updates

	listener := agent.New(
		agent.WithPersister(persister),
		agent.WithAdapter(agent.NewSSEAdapter("http://localhost:8080/sse?project=project-1&apikey=apikey-1111111111")),
		agent.WithSubscriber(subscriber),
	)

	go func() {
		for resource := range subscriber {
			jsonMsg, _ := json.Marshal(resource)
			slog.Info("Received update", "resource", jsonMsg)
			// Process the resource update as needed
		}
	}()

	if err := listener.Run(context.Background()); err != nil {
		panic(err)
	}

	log.Println("Agent is running...")
}
