# Sentinel

A monitoring and data management system built with Go.

## Components

### Resagent (WIP)

A Go library for interacting with [Resgate](./resgate) servers, providing caching and real-time data synchronization.

**Features:**
- Cache management with file-based storage
- Real-time sync via Server-Sent Events
- Resource change subscriptions
- Resgate protocol compatibility

**Progress Checklist:**
- [ ] Core interfaces definition (Cache, Client, Adapter)
- [ ] File-based cache implementation
- [ ] HTTP client for Resgate API
- [ ] SSE adapter for real-time updates
- [ ] Subscription management system
- [ ] Observer pattern implementation
- [ ] Configuration and options
- [ ] Error handling and reconnection logic
- [ ] Unit tests
- [ ] Integration tests with Resgate
- [ ] Documentation and examples

**Usage:**
```go
agent := resagent.New(
    resagent.WithResgateURL("http://localhost:8080"))
agent.Start()
data, err := agent.Get("user.123")
subscription := agent.Subscribe("user.*", handleChange)
```

### Resgate

Real-time API Gateway that enables real-time web APIs using the RES protocol.

**Features:**
- Resource-based API structure (RID patterns)
- Event-driven data synchronization
- Client subscription management

**Progress Checklist:**
- [x] Basic project structure
- [x] HTTP client implementation
- [x] NATS adapter integration
- [x] SSE adapter for real-time events
- [x] Authorization with Casbin
- [x] File-based resource persistence
- [ ] Performance benchmarks
- [ ] Load testing scenarios


## Getting Started

```bash
git clone https://github.com/lamlv2305/sentinel.git
cd sentinel/resagent
go mod tidy
```