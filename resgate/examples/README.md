# ResGate SSE Example

This example demonstrates a complete ResGate Server-Sent Events implementation with sample data generation.

## Features

- **HTTP Server**: Runs on port 9005 with health check endpoint
- **ResGate Integration**: Full configuration with persister, authorizer, and SSE adapter
- **SSE Endpoint**: Available at `/sse` with credential validation
- **Sample Events**: Automatically generates random resource change events every 5 seconds
- **Graceful Shutdown**: Handles SIGINT/SIGTERM signals properly
- **Web Client**: HTML client for testing SSE connections

## Prerequisites

The example includes all necessary files:

- `rbac_model.conf` - Casbin RBAC model configuration
- `server.go` - Main server implementation
- `client.html` - Web client for testing
- `README.md` - This documentation

The data directory will be created automatically for file persistence.

## Running the Example

1. **Start the server**:

   ```bash
   go run server.go
   ```

2. **Test the health endpoint**:

   ```bash
   curl http://localhost:9005/health
   ```

3. **Test SSE connection**:

   ```bash
   curl -N "http://localhost:9005/sse?apikey=sample-api-key-12345&project=project-alpha"
   ```

4. **Use the web client**:
   - Open `client.html` in your browser
   - Click "Connect" to start receiving events
   - Watch real-time events in the log

## Sample Events

The server generates random events every 5 seconds with:

- **Projects**: project-alpha, project-beta, project-gamma
- **Groups**: users, documents, settings, notifications
- **Resource Types**: text, json_object, json_array
- **Actions**: create, update, delete

## SSE Endpoint Parameters

- `apikey`: Must be at least 10 characters (simple validation)
- `project`: Project identifier (project-alpha, project-beta, project-gamma)

## Example SSE Output

```
data: {"type":"connected","id":"a1b2c3d4-e5f6-7890-abcd-ef1234567890"}

data: eyJhY3Rpb24iOiJjcmVhdGUiLCJ0aW1lc3RhbXAiOiIyMDI1LTA3LTE3VDEwOjMwOjAwWiIsInJlc291cmNlIjp7InJlc291cmNlX2lkIjoicmVzb3VyY2UtMSIsInByb2plY3RfaWQiOiJwcm9qZWN0LWFscGhhIiwiZ3JvdXAiOiJ1c2VycyIsInJlc291cmNlX3R5cGUiOiJ0ZXh0IiwiZGF0YSI6IlNhbXBsZSB0ZXh0IGRhdGEgZm9yIGV2ZW50IDEgLSAxMDozMDowMCJ9fQ==

: keepalive
```

## Web Client Features

- **Connection Management**: Connect/disconnect with different projects
- **Real-time Stats**: Event count, connection time, last event time
- **Event Log**: Formatted display of all received events
- **Base64 Decoding**: Automatically decodes and formats JSON events
- **Multiple Projects**: Switch between different project streams

## Architecture

```
HTTP Server (port 9005)
├── /health          - Health check endpoint
├── /sse             - Server-Sent Events endpoint
└── Static Files     - Could serve client.html

ResGate Components:
├── File Persister   - Stores data in ./data directory
├── Casbin Authorizer - RBAC with rbac_model.conf
└── SSE Adapter      - Handles real-time event streaming

Sample Generator:
└── Background task generating events every 5 seconds
```

## Customization

You can modify the example to:

1. **Change event frequency**: Modify the ticker duration in `generateSampleEvents()`
2. **Add more projects**: Extend the `projectIds` slice
3. **Custom data**: Modify `generateSampleData()` function
4. **Authentication**: Enhance the `credentialValidator` function
5. **Persistence**: The file persister saves data that can be queried later

## Troubleshooting

1. **Port already in use**: Change the port in both server.go and client.html
2. **CORS issues**: The server includes CORS headers for web clients
3. **Connection drops**: Check the keepalive interval (30 seconds by default)
4. **No events**: Verify the sample generator is running and check server logs

## Production Considerations

For production use, consider:

1. **Authentication**: Implement proper JWT or OAuth validation
2. **Rate Limiting**: Add rate limiting for connections and events
3. **Monitoring**: Add metrics and health checks
4. **Persistence**: Use a proper database instead of file storage
5. **Clustering**: Handle multiple server instances
6. **SSL/TLS**: Use HTTPS for secure connections
