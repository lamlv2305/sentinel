package adapter

import (
	"log/slog"
	"sync"
	"time"
)

type hub struct {
	mu      *sync.RWMutex
	clients map[string]map[string]*Client
	logger  *slog.Logger
}

func defaultHub() *hub {
	return &hub{
		mu:      &sync.RWMutex{},
		clients: make(map[string]map[string]*Client),
		logger:  slog.Default(),
	}
}

func (d *hub) add(c *Client) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.clients[c.ProjectId]; !ok {
		d.clients[c.ProjectId] = make(map[string]*Client)
	}

	d.clients[c.ProjectId][c.Id] = c
}

func (d *hub) remove(projectId, clientId string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	clients, ok := d.clients[projectId]
	if !ok {
		return
	}

	client, exists := clients[clientId]
	if !exists {
		return
	}

	client.Close()
	delete(clients, clientId)

	if len(clients) == 0 {
		delete(d.clients, projectId)
	}
}

func (d *hub) broadcast(projectId string, message string) {
	d.mu.RLock()
	clients, ok := d.clients[projectId]
	if !ok {
		d.mu.RUnlock()
		return
	}

	// Send messages concurrently without copying the entire client list
	var wg sync.WaitGroup
	for _, client := range clients {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			if err := c.SendWithTimeout(message, 5*time.Second); err != nil {
				d.logger.Warn("Client send failed, removing client",
					"projectId", projectId,
					"clientId", c.Id,
					"error", err)
				// Remove client in a separate goroutine to avoid deadlock
				go d.remove(projectId, c.Id)
			}
		}(client)
	}
	d.mu.RUnlock()
	wg.Wait()
}

// healthCheck performs a health check on all clients and removes dead ones
func (d *hub) cleanup() {
	d.mu.RLock()
	var disconnectedClients []struct{ projectId, clientId string }

	for projectId, clients := range d.clients {
		for clientId, client := range clients {
			if !client.IsConnected() {
				disconnectedClients = append(disconnectedClients,
					struct{ projectId, clientId string }{projectId: projectId, clientId: clientId})
			}
		}
	}
	d.mu.RUnlock()

	// Remove disconnected clients
	for _, dc := range disconnectedClients {
		d.logger.Info("Removing disconnected client",
			"projectId", dc.projectId,
			"clientId", dc.clientId)
		d.remove(dc.projectId, dc.clientId)
	}
}
