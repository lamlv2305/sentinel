# Sentinel

A monitoring and data management system built with Go.

## Components

### Resagent (WIP)

A Go library for interacting with [Resgate](https://resgate.io/) servers, providing caching and real-time data synchronization.

**Features:**
- Cache management with file-based storage
- Real-time sync via Server-Sent Events
- Resource change subscriptions
- Resgate protocol compatibility

**Usage:**
```go
agent := resagent.New(
    resagent.WithResgateURL("http://localhost:8080"),
    resagent.WithCacheDir("./cache"),
)

agent.Start()
data, err := agent.Get("user.123")
subscription := agent.Subscribe("user.*", handleChange)
```

### Resgate

Real-time API Gateway that enables real-time web APIs.

## Getting Started

```bash
git clone https://github.com/lamlv2305/sentinel.git
cd sentinel/resagent