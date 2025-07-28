package operator

import (
	"context"
	"sync"
	"time"
)

type Client struct {
	Id        string
	ProjectId string

	ch   chan string
	done chan struct{}
	mu   sync.Mutex
}

func NewClient(id, projectId string) *Client {
	return &Client{
		Id:        id,
		ProjectId: projectId,
		ch:        make(chan string, 100), // Larger buffer to handle bursts
		done:      make(chan struct{}),
	}
}

// Close gracefully closes the client
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.done:
		// Already closed
		return
	default:
		close(c.done)
		close(c.ch)
	}
}

// IsConnected checks if the client is still connected
func (c *Client) IsConnected() bool {
	select {
	case <-c.done:
		return false
	default:
		return true
	}
}

// GetChannel returns the client's message channel for reading
func (c *Client) GetChannel() chan string {
	return c.ch
}

// SendWithTimeout sends a message with a timeout
func (c *Client) SendWithTimeout(message string, timeout time.Duration) error {
	if !c.IsConnected() {
		return context.Canceled
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return ctx.Err()

	case c.ch <- message:
		return nil

	case <-c.done:
		return context.Canceled
	}
}
