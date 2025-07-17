package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/lamlv2305/sentinel/resgate"
	"github.com/lamlv2305/sentinel/resgate/adapter"
	"github.com/lamlv2305/sentinel/resgate/resource"
	"github.com/lamlv2305/sentinel/types"
)

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create HTTP server mux
	mux := http.NewServeMux()

	// Setup health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create file-based persister
	persister := resource.NewFilePersister("./data")

	// Create Casbin authorizer
	authorizer, err := resource.NewCasbinAuthorizer(nil, "./rbac_model.conf")
	if err != nil {
		logger.Error("Failed to create authorizer", "error", err)
		os.Exit(1)
	}

	// Create credential validator
	credentialValidator := func(ctx context.Context, apikey, project string) error {
		// Simple validation - in production use proper authentication
		if apikey == "" {
			return fmt.Errorf("missing apikey")
		}
		if project == "" {
			return fmt.Errorf("missing project")
		}
		if len(apikey) < 10 {
			return fmt.Errorf("invalid apikey")
		}
		logger.Info("Client authenticated", "project", project, "apikey", apikey[:5]+"...")
		return nil
	}

	// Create SSE adapter
	sseAdapter := adapter.NewSSE(mux, "/sse", credentialValidator)

	// Create ResGate with all possible options
	rg := resgate.NewResGate(
		resgate.WithLogger(logger),
		resgate.WithPersister(persister),
		resgate.WithAuthorizer(authorizer),
		resgate.WithAdapter(sseAdapter),
	)

	// Setup sample data generator
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start sample event generator
	go generateSampleEvents(ctx, sseAdapter, logger)

	// Start HTTP server
	server := &http.Server{
		Addr:    ":9005",
		Handler: mux,
	}

	// Start server in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting HTTP server", "port", 9005)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
		}
	}()

	// Start ResGate adapter
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting ResGate adapter")
		if err := rg.Run(ctx); err != nil {
			logger.Error("ResGate error", "error", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down server...")

	// Cancel context to stop background tasks
	cancel()

	// Shutdown HTTP server gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", "error", err)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	logger.Info("Server stopped")
}

// generateSampleEvents creates random resource change events every 5 seconds
func generateSampleEvents(ctx context.Context, adapter *adapter.SSE, logger *slog.Logger) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	projectIds := []string{"project-alpha", "project-beta", "project-gamma"}
	groups := []string{"users", "documents", "settings", "notifications"}
	resourceTypes := []types.ResourceType{
		types.ResourceTypeText,
		types.ResourceTypeJsonObject,
		types.ResourceTypeJsonArray,
	}
	actionTypes := []types.ActionType{
		types.ActionTypeCreate,
		types.ActionTypeUpdate,
		types.ActionTypeDelete,
	}

	eventCount := 0

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping sample event generator")
			return
		case <-ticker.C:
			eventCount++

			// Generate random event
			projectId := projectIds[rand.Intn(len(projectIds))]
			group := groups[rand.Intn(len(groups))]
			resourceType := resourceTypes[rand.Intn(len(resourceTypes))]
			action := actionTypes[rand.Intn(len(actionTypes))]

			event := types.ChangedEvent{
				Action:    action,
				Timestamp: time.Now(),
				Resource: types.Resource{
					ResourceId:   fmt.Sprintf("resource-%d", eventCount),
					ProjectId:    projectId,
					Group:        group,
					ResourceType: resourceType,
					Data:         generateSampleData(resourceType, eventCount),
				},
			}

			// Send event through adapter
			if err := adapter.OnChanged(ctx, event); err != nil {
				logger.Error("Failed to send event", "error", err)
			} else {
				logger.Info("Generated sample event",
					"eventId", eventCount,
					"action", action,
					"projectId", projectId,
					"group", group,
					"resourceId", event.Resource.ResourceId,
					"resourceType", resourceType)
			}
		}
	}
}

// generateSampleData creates sample data based on resource type
func generateSampleData(resourceType types.ResourceType, eventCount int) []byte {
	switch resourceType {
	case types.ResourceTypeText:
		return []byte(fmt.Sprintf("Sample text data for event %d - %s",
			eventCount, time.Now().Format("15:04:05")))

	case types.ResourceTypeJsonObject:
		return []byte(fmt.Sprintf(`{
			"eventId": %d,
			"timestamp": "%s",
			"status": "active",
			"value": %d,
			"metadata": {
				"source": "sample-generator",
				"version": "1.0"
			}
		}`, eventCount, time.Now().Format(time.RFC3339), rand.Intn(1000)))

	case types.ResourceTypeJsonArray:
		return []byte(fmt.Sprintf(`[
			{"id": %d, "name": "Item %d", "active": true},
			{"id": %d, "name": "Item %d", "active": false},
			{"id": %d, "name": "Item %d", "active": true}
		]`, eventCount, eventCount, eventCount+1, eventCount+1, eventCount+2, eventCount+2))

	default:
		return []byte(fmt.Sprintf("Binary data for event %d", eventCount))
	}
}
