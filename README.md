# Order State Machine

A small Go service that models Gametime-style checkout order lifecycle with enforced state transitions, stage-dependent failure recovery, and explicit handling of partial failures.

## What I Built and Why

Checkout orders move through distinct stages — each with different recovery rules when something fails. A payment decline before authorization needs no cleanup; a completion failure after authorization requires voiding the charge; if the void also fails, the order must land in a state that demands manual resolution rather than pretending everything is fine.

The core design separates three concerns:

1. **State machine** (internal/order/machine.go) — defines valid transitions and records timestamped history on every change.
2. **Service orchestration** (internal/order/service.go) — owns stage-dependent recovery: authorize decline → ejected; complete fail → void → cancelled or 
eeds_attention.
3. **Payment stub** (internal/payment/payment.go) — interface boundary so tests can inject failures without a real payment processor.

The API uses explicit endpoints (/authorize, /complete) rather than a generic "advance" action so intent is clear in both code and HTTP logs.

### State Diagram

`
initialized ──authorize OK──► payment_authorized ──complete OK──► complete
     │                                │
     │ authorize declined             ├──complete fail + void OK──► cancelled
     ▼                                └──complete fail + void fail──► needs_attention
  rejected
`

## How to Run

**Requirements:** Go 1.22+

`ash
# Run tests
go test ./...

# Start the server (default :8080)
go run ./cmd/server

# Optional: simulate payment failures via env vars
PAYMENT_AUTHORIZE_FAIL=true go run ./cmd/server   # all authorizations decline
COMPLETE_FAIL=true go run ./cmd/server            # all completions fail
VOID_FAIL=true go run ./cmd/server                # all voids fail (use with COMPLETE_FAIL)
`

### Example API Flow (happy path)

`ash
# Create order
curl -s -X POST http://localhost:8080/api/v1/orders | jq

# Authorize payment (replace {id} with order id from above)
curl -s -X POST http://localhost:8080/api/v1/orders/{id}/authorize | jq

# Complete order
curl -s -X POST http://localhost:8080/api/v1/orders/{id}/complete | jq

# Query state and history
curl -s http://localhost:8080/api/v1/orders/{id} | jq
`

### API Reference

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/v1/orders | Create a new order in initialized state |
| POST | /api/v1/orders/{id}/authorize | Authorize payment |
| POST | /api/v1/orders/{id}/complete | Complete the order |
| GET | /api/v1/orders/{id} | Get current state and full history |

## Tradeoffs

- **In-memory storage** — simple and sufficient for the prototype; data is lost on restart.
- **Synchronous void on failure** — keeps the failure path easy to reason about, but in production you'd likely retry voids asynchronously with backoff.
- **No idempotency keys** — duplicate API calls could attempt transitions twice; production would need idempotent handlers and deduplication.
- **Global payment stub via env** — convenient for manual demos; per-order failure injection would be needed for richer testing via HTTP.
- **No authentication** — appropriate for a take-home; a real service would gate these endpoints.

## With More Time

- **Persistence** — Postgres with optimistic locking on order rows; outbox pattern for side effects.
- **Async void retries** — queue failed voids with exponential backoff and a dead-letter queue; alert on exhaustion.
- **Observability** — metrics on 
eeds_attention count, structured logging, tracing through payment calls.
- **Admin resolution API** — endpoint for ops to manually resolve 
eeds_attention orders after investigating.
- **Idempotency** — accept Idempotency-Key header on mutating endpoints; store results for replay.
- **OpenAPI spec** — document the API for client generation and review.
