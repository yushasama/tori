# tori

A fast Shopify monitor written in Go.

Built as a backend experiment to explore HFT-style infrastructure design in a lightweight context. While it monitors Shopify endpoints, the core focus is on lock-free queuing, memory hygiene, and sliding window rate-limited dispatching.

## Features

* **Parallel Monitor Engine** - One goroutine per `site x monitor`, cancel-safe and backoff-aware.
* **Bitwise Capacity Rounding** - `rateLimit` is rounded to the nearest power of two for fast masking.
* **Lock-Free Ring Buffer** - Circular queues with bitmask indexing for both job queue and dispatch timestamps.
* **Sliding Window Rate Limiter** - Timestamped ring buffer prunes old sends and enforces N-per-interval dispatch policy.
* **Dispatcher with Backpressure** - Centralized dispatcher throttles sends based on past timestamps and queue fullness.
* **GC-Free Hot Path** - Avoids heap allocations in polling and dispatching via reused structs and preallocated slices.
* **Retry Semantics** - Failed polls back off and retry up to N times per monitor (configurable).
* **Goroutine Hygiene** - All goroutines respond to context cancellation and are cleanly terminated.
* **Built-In Profiling** - pprof endpoints and `GODEBUG=gctrace=1` enabled for live system inspection.
* **YAML-Based Config** - Fully extensible via `config.yaml`.

## Config

Monitors and Discord webhooks are defined in `config.yaml`. Each monitor instance is launched in parallel and maintains its own polling loop, with centralized webhook dispatching via the dispatcher. No DNS fallback or proxy rotation is currently implemented. If an endpoint becomes unreachable due to a DNS issue or connection failure, it will simply retry after backoff.

## Design Highlights

### Dispatcher and Rate-Limited Scheduling

The dispatcher receives jobs from all monitors and sends them via Discord webhooks. A fixed-size timestamp ring tracks the most recent N sends within a moving window:

* The buffer is pruned every tick to drop timestamps older than the interval.
* If the number of remaining timestamps is below `rateLimit`, a new job is dispatched.
* Otherwise, the job waits in the pending queue (which is also a ring buffer).

This implements a **sliding window LRU rate limiter** — allowing bursty but bounded throughput without violating webhook limits.

### Lock-Free Ring Buffers

Both `pending` jobs and `sent` timestamps use circular queues with power-of-two sizes:

```go
size := 1 << bits.Len(uint(rateLimit - 1))
index := pos & (size - 1)
```

This avoids modulus overhead, improves cache locality, and reduces allocator pressure.

### Backpressure and Job Eviction

If the job queue is full, the dispatcher drops the oldest job (LRU-style eviction). This ensures memory bounds are respected under load and prevents slow endpoints from stalling the system.

Crucially, the system prioritizes the **freshest jobs** by always dispatching the most recent entries first. This mirrors "alpha decay" in financial systems - the insight or opportunity embedded in the job (e.g. a restock event) rapidly loses value over time. Dispatching old jobs after a delay would result in notifying users about stale events that may have already sold out.

In latency-sensitive environments such as sneaker bots, low-volume product launches, or arbitrage monitors, freshness is paramount. This design ensures the system remains responsive to **real-time state** rather than delivering outdated intelligence.

### Retry and Fault Tolerance

* Polling failures are retried with exponential backoff.
* DNS-level failures (e.g. domain unreachable or DNS timeout) are not recoverable. The system lacks proxy rotation, dynamic DNS resolution, or failover pathing.
* Webhook delivery uses a single endpoint per monitor and does not support failover to backup URLs or retry persistence (e.g. job replay on failure).

## Memory and Concurrency Guarantees

| Concern             | Approach                                                                     |
| ------------------- | ---------------------------------------------------------------------------- |
| GC Pressure         | Preallocated ring buffers, reused job structs, and zero allocation hot paths |
| Heap Churn          | No boxed values in fast paths; fixed-capacity slices reused throughout       |
| Goroutine Lifecycle | Full context cancellation propagation, bounded retry logic                   |
| Dispatch Throttling | Sliding window timestamp ring buffer with pruning and masking                |
| Queue Safety        | Lock-free, bitmask-indexed ring queues for both jobs and send history        |
| Backpressure        | Oldest job dropped if pending buffer exceeds capacity                        |
| Retry Semantics     | Configurable polling backoff and retry count per monitor                     |
| Shutdown Hygiene    | All goroutines respect cancellation and clean up on shutdown                 |

## Usage

```bash
go build -o tori cmd/main.go
./tori
```

To inspect profiling data:

```bash
go tool pprof -http=:9999 tori.exe profiles/CPU_<timestamp>.prof
```

Enable GC tracing:

```bash
GODEBUG=gctrace=1 ./tori
```

## Caveats

* No proxy rotation or DNS failure recovery
* No webhook failover or retry queue for delivery
* Built as a learning tool, not hardened for production use

## Purpose

`tori` is a backend systems sandbox. It demonstrates:

* How to build bounded, concurrent queues without locks
* Designing rate-limited systems using sliding window semantics
* Structuring a system for high-throughput, low-GC dispatching
* Goroutine lifecycle management under real-world constraints

This project was built for the intent of exploring advanced backend infrastructure, performance profiling & memory-safe concurrency design in Go.
