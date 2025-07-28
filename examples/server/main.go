package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/lamlv2305/sentinel/operator"
	"github.com/lamlv2305/sentinel/types"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Allow specific headers
		w.Header().Set("Access-Control-Allow-Headers", "*")
		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	slog.Debug("Initializing SSE server...", "address", ":8080/sse")
	sse := operator.NewSSE(
		mux,
		"/sse",
		operator.WithSSECredentialVerifier(func(ctx context.Context, apikey string, project string) error {
			// Implement your credential verification logic here
			if apikey == "" || project == "" {
				return errors.New("invalid credentials")
			}
			return nil
		}),
	)
	go sse.Run(context.Background())

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			encoded := []byte(time.Now().String())

			slog.Debug("Broadcasting event", "data", encoded)

			err := sse.Broadcast(context.Background(), types.ChangedEvent{
				Action:    types.ActionTypeUpdate,
				Timestamp: time.Now(),
				Resource: types.Resource{
					ResourceId:   "resource-1",
					ProjectId:    "project-1",
					Group:        "",
					ResourceType: types.ResourceTypeText,
					Data:         encoded,
				},
			})
			if err != nil {
				slog.Error("Failed to broadcast event", "error", err)
			}
		}
	}()

	slog.Debug("Starting SSE server on :8080")
	if err := http.ListenAndServe(":8080", withCORS(mux)); err != nil {
		panic(err)
	}
}
