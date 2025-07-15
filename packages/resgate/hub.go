package resgate

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

	if clients, ok := d.clients[projectId]; ok {
		if client, exists := clients[clientId]; exists {
			client.Close() // Properly close the client
			delete(clients, clientId)

			if len(clients) == 0 {
				delete(d.clients, projectId)
			}
		}
	}
}

func (d *hub) broadcast(projectId string, message string) {
	d.mu.RLock()
	clients, ok := d.clients[projectId]
	if !ok {
		d.mu.RUnlock()
		return
	}

	// Create a copy of clients to avoid holding the lock during sending
	clientList := make([]*Client, 0, len(clients))
	for _, client := range clients {
		clientList = append(clientList, client)
	}
	d.mu.RUnlock()

	// Send messages concurrently to avoid blocking
	var wg sync.WaitGroup
	for _, client := range clientList {
		wg.Add(1)
		go func(c *Client) {
			defer wg.Done()
			d.sendToClient(c, message, projectId)
		}(client)
	}
	wg.Wait()
}

// sendToClient sends a message to a specific client with timeout and error handling
func (d *hub) sendToClient(client *Client, message string, projectId string) {
	if err := client.SendWithTimeout(message, 5*time.Second); err != nil {
		// Client send failed - remove the client
		d.logger.Warn("Client send failed, removing client",
			"projectId", projectId,
			"clientId", client.Id,
			"error", err)
		d.remove(projectId, client.Id)
		client.Close() // Ensure client is properly closed
	}
}

// getClientCount returns the number of clients for a specific project
func (d *hub) getClientCount(projectId string) int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if clients, ok := d.clients[projectId]; ok {
		return len(clients)
	}
	return 0
}

// getTotalClientCount returns the total number of clients across all projects
func (d *hub) getTotalClientCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	total := 0
	for _, clients := range d.clients {
		total += len(clients)
	}
	return total
}

// getProjectIds returns a list of all project IDs that have active clients
func (d *hub) getProjectIds() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	projectIds := make([]string, 0, len(d.clients))
	for projectId := range d.clients {
		projectIds = append(projectIds, projectId)
	}
	return projectIds
}

// healthCheck performs a health check on all clients and removes dead ones
func (d *hub) healthCheck() {
	d.mu.RLock()
	allClients := make(map[string]map[string]*Client)
	for projectId, clients := range d.clients {
		allClients[projectId] = make(map[string]*Client)
		for clientId, client := range clients {
			allClients[projectId][clientId] = client
		}
	}
	d.mu.RUnlock()

	for projectId, clients := range allClients {
		for clientId, client := range clients {
			// Check if client is still connected
			if !client.IsConnected() {
				d.logger.Info("Removing disconnected client",
					"projectId", projectId,
					"clientId", clientId)
				d.remove(projectId, clientId)
				continue
			}

			// Check if client is stale (no activity for 5 minutes)
			if time.Since(client.GetLastSeen()) > 5*time.Minute {
				d.logger.Info("Removing stale client",
					"projectId", projectId,
					"clientId", clientId,
					"lastSeen", client.GetLastSeen())
				d.remove(projectId, clientId)
			}
		}
	}
}
