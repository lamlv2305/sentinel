package operator

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/lamlv2305/sentinel/types"
)

var _ Adapter = &SSE{}

type WithSSE func(*SSE)

func WithSSECredentialVerifier(cv CredentialVerifier) WithSSE {
	return func(s *SSE) {
		s.cv = cv
	}
}

func WithOnConnectedHook(hook func(ctx context.Context, client *Client)) WithSSE {
	return func(s *SSE) {
		s.hook.OnConnected = append(s.hook.OnConnected, hook)
	}
}

func WithOnDisconnectedHook(hook func(ctx context.Context, client *Client)) WithSSE {
	return func(s *SSE) {
		s.hook.OnDisconnected = append(s.hook.OnDisconnected, hook)
	}
}

type SSE struct {
	mux      *http.ServeMux
	endpoint string
	hub      *hub
	cv       CredentialVerifier
	hook     Hook
}

func NewSSE(mux *http.ServeMux, endpoint string, opts ...WithSSE) *SSE {
	ins := &SSE{
		mux:      mux,
		endpoint: endpoint,
		hub:      defaultHub(),
		cv:       nil,
		hook:     Hook{},
	}

	for _, opt := range opts {
		opt(ins)
	}

	if ins.cv == nil {
		ins.cv = func(ctx context.Context, apikey string, project string) error {
			slog.Error("Credential verifier not set")
			return errors.New("credential verifier not set")
		}
	}

	return ins
}

func (s *SSE) Run(ctx context.Context) error {
	s.mux.HandleFunc(s.endpoint, s.OnConnected)

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			s.hub.cleanup() // Clean up disconnected clients
		}
	}
}

// OnChanged implements Adapter.
func (s *SSE) Broadcast(ctx context.Context, event types.ChangedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	message := "data: " + base64.StdEncoding.EncodeToString(data)
	s.hub.broadcast(event.Resource.ProjectId, message)
	return nil
}

func (s *SSE) OnConnected(w http.ResponseWriter, r *http.Request) {
	// Handle panics gracefully
	defer func() {
		if err := recover(); err != nil {
			slog.Error("Recovered from panic in OnConnected", err)
		}
	}()

	// Validate credentials
	apikey := r.URL.Query().Get("apikey")
	project := r.URL.Query().Get("project")

	if err := s.cv(r.Context(), apikey, project); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Setup SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Flush headers immediately
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	flusher.Flush()

	// Create and register client
	connectionId := uuid.New().String()
	client := NewClient(connectionId, project)
	s.hub.add(client)
	defer func() {
		client.Close()
		s.hub.remove(project, connectionId)
	}()

	// Send connection confirmation
	connectionEvent := "data: {\"type\":\"connected\",\"id\":\"" + connectionId + "\"}\n\n"
	if _, err := w.Write([]byte(connectionEvent)); err != nil {
		return
	}
	flusher.Flush()

	// Start event loop
	s.handleEvents(w, r, client, flusher)
}

// handleEvents manages the SSE event loop for a connected client
func (s *SSE) handleEvents(w http.ResponseWriter, r *http.Request, client *Client, flusher http.Flusher) {
	clientCh := client.GetChannel()
	keepalive := time.NewTicker(30 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case message, ok := <-clientCh:
			if !ok {
				return
			}
			if s.writeSSE(w, message+"\n\n", flusher) != nil {
				return
			}
		case <-keepalive.C:
			if s.writeSSE(w, ": keepalive\n\n", flusher) != nil {
				return
			}
		}
	}
}

// writeSSE writes SSE data and flushes
func (s *SSE) writeSSE(w http.ResponseWriter, data string, flusher http.Flusher) error {
	if _, err := w.Write([]byte(data)); err != nil {
		return err
	}
	flusher.Flush()

	return nil
}
