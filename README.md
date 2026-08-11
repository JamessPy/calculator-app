# Calculator

Go REST API + React/TypeScript frontend. No third-party backend dependencies.

## Setup

**Docker** (both services):

```bash
docker compose up --build
```

Frontend on http://localhost:5173, API on http://localhost:8080.

**Local** — requires Go 1.26+ and Node 20+:

```bash
go run ./cmd/api          # API on :8080 (override with PORT)

cd web && npm install
npm run dev               # frontend on :5173
```

## API

`POST /api/v1/calculate` — `add`, `subtract`, `multiply`, `divide`, `power`, `sqrt`, `percentage`. Operation names are case-insensitive; `b` is omitted for `sqrt`.

```bash
curl -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"add","a":2,"b":3}'
# {"operation":"add","a":2,"b":3,"result":5}

curl -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"sqrt","a":144}'
# {"operation":"sqrt","a":144,"result":12}

curl -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"divide","a":10,"b":0}'
# 422 {"error":{"code":"DIVISION_BY_ZERO","message":"division by zero is undefined"}}
```

**400** — malformed request: `INVALID_JSON`, `VALIDATION_FAILED`, `UNSUPPORTED_OPERATION`, `OPERAND_REQUIRED`, `OPERAND_NOT_FINITE`
**422** — valid request, undefined mathematics: `DIVISION_BY_ZERO`, `NEGATIVE_SQUARE_ROOT`, `RESULT_NOT_FINITE`
**500** — `INTERNAL_ERROR`; details are logged, never returned

`GET /healthz` → `{"status":"ok"}`

## Tests

```bash
go test ./internal/...              # backend
cd web && npm test                  # frontend
```

Coverage: backend **99.1%**, frontend **98.8%**. Reports in [`docs/`](./docs) — regenerate with `go tool cover -html` and `npm run coverage`.

## Design decisions

**One service, not several.** The brief asks for "a backend microservice" —
singular — so the goal here is one service built to microservice standards
rather than a distributed system. It has a single responsibility, holds no
state, is independently deployable as a container, is configured through the
environment, exposes a health endpoint, and versions its API. Splitting
arithmetic into `add-service`, `divide-service` and a gateway would look more
like "microservices" while being worse in every way that matters: three
network hops, three deployment pipelines, and no independent scaling or data
ownership to justify any of it. Service boundaries follow business
capabilities, not functions. If this service later grew a stateful capability
— calculation history, say — that would be a genuine seam: a separate database,
a different scaling profile, and an asynchronous event between the two.

**Layout** follows the [official Go module guidance](https://go.dev/doc/modules/layout) for server projects: the binary in `cmd/`, all application code in `internal/`.

**One endpoint** with the operation in the request body, rather than seven endpoints repeating the same validation.

**The domain layer knows nothing about HTTP.** `internal/calculator` imports only `errors` and `math`, so the arithmetic is testable without a server; `internal/httpapi` owns status codes, JSON and CORS. Go's ban on import cycles enforces the direction.

**400 vs 422** separates "the request is wrong" from "the request is fine but the maths is undefined", so the frontend can tell a bug from a user error. Errors carry a stable `code` and the frontend supplies its own wording, so backend messages can change without breaking the client.

**Optional operands are pointers** (`*float64`, `omitempty`, `b?:`) because `0` is a legitimate operand and must not be confused with "not provided" — `10 / 0` has to reach the division-by-zero check.

**Non-finite results are rejected.** `1e308 * 10` is `+Inf`, which JSON cannot encode; returning it would make `encoding/json` fail after the 200 status was already written.

**Rounding happens in the UI.** The API returns the exact `float64` (`0.1 + 0.2` → `0.30000000000000004`); the display applies `toPrecision(12)`.

**The domain is strict, the transport forgiving.** `ToLower`/`TrimSpace` run in the handler, so the domain rule stays in one place.

**CORS is an allow-list**, not a reflected `Origin` header.

**Standard library only.** `net/http.ServeMux` handles routing (method patterns since Go 1.22 give 404/405 for free) and `log/slog` handles structured logging, so `go.mod` has no dependencies.
