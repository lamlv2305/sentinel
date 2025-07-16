# Sentinel

A monitoring and data management system built with Go.

## Components

### Resagent (WIP)

A Go library for interacting with [Resgate](./resgate) servers, providing caching and real-time data synchronization.

**Features:**
- Cache management with file-based storage
- Real-time sync
- Resource change subscriptions

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


## Getting Started

```bash
git clone https://github.com/lamlv2305/sentinel.git
cd sentinel/resagent
go mod tidy
```