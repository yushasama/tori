# tori

A fast Shopify monitor written in Go.

This was built to explore high-performance infrastructure concepts in a lightweight setting. It features:

- A custom ring buffer for queueing payloads without locks
- Decoupled polling and dispatch logic
- Graceful handling of Discord's rate limits
- Basic GC and pprof tracing for runtime introspection

Currently, it does not support proxies, but it **can monitor multiple stores in parallel** via goroutines.

## Features

- Shopify monitor in Golang
- Lock-free ring buffer with cache-conscious structure
- Dispatcher queues to comply with rate-limited APIs like Discord
- JSON-configurable monitors for Shopify endpoints
- Built-in pprof and `GODEBUG=gctrace=1` support for performance visibility

## Usage

```bash
go build -o tori cmd/main.go
./tori
```

To view performance profiling (CPU, heap):
```bash
go tool pprof -http=:9999 tori.exe profiles/CPU_<timestamp>.prof
```

Enable GC tracing:
```bash
GODEBUG=gctrace=1 ./tori
```

## Notes
This project is a backend experiment, not a polished production tool.

Built to test queuing systems, API dispatching, and Go's memory/runtime behavior under load.
