package resgate

import (
	"context"
	"sync"
	"time"
)

type Client struct {
	Id        string
	ProjectId string

	ch       chan string
	done     chan struct{}
	mu       sync.RWMutex
	lastSeen time.Time
}

func NewClient(id, projectId string) *Client {
	return &Client{
		Id:        id,
		ProjectId: projectId,
		ch:        make(chan string, 100), // Larger buffer to handle bursts
		done:      make(chan struct{}),
		lastSeen:  time.Now(),
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

// UpdateLastSeen updates the last seen timestamp
func (c *Client) UpdateLastSeen() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastSeen = time.Now()
}

// GetLastSeen returns the last seen timestamp
func (c *Client) GetLastSeen() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastSeen
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

// SendWithTimeout sends a message with a timeout
func (c *Client) SendWithTimeout(message string, timeout time.Duration) error {
	if !c.IsConnected() {
		return context.Canceled
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case c.ch <- message:
		c.UpdateLastSeen()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return context.Canceled
	}
}
