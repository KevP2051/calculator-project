# Calculator Backend

A stateless REST microservice for arithmetic, written in Go with no third-party
dependencies. Seven operations over one endpoint, with a strict error contract: every
failure returns the same JSON envelope carrying a stable machine-readable code.

Built spec-first with [Spec Kit](https://github.com/github/spec-kit). The specification,
plan, research notes and task list live in
[`specs/001-calculator-rest-api/`](../specs/001-calculator-rest-api/).

---

## Prerequisites

**Go 1.22 or later.** The floor is load-bearing: the service routes with the standard
library's `ServeMux` using method patterns (`"POST /api/v1/calculate"`), which Go 1.22
introduced. That is what lets the project avoid a router dependency.

```bash
go version
```

No database, no message broker, no container runtime.

---

## Run the server

From this directory:

```bash
go run ./cmd/server
```

Listens on `:8080`. Confirm it is up:

```bash
curl -s http://localhost:8080/healthz
```

```json
{"status":"ok"}
```

To build a binary instead:

```bash
go build -o server ./cmd/server && ./server
```

### Configuration

Two environment variables, both optional:

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | Listen port |
| `CORS_ORIGIN` | `http://localhost:5173` | Single allowed browser origin. The default is the Vite dev server. |

```bash
PORT=9000 CORS_ORIGIN=http://localhost:3000 go run ./cmd/server
```

Nothing else is configurable, by design.

---

## Run the tests

```bash
go test ./... -cover
```

Coverage report:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

Current coverage — regenerate with the commands above; the report itself is gitignored:

| Package | Coverage |
|---|---|
| `internal/calc` | 100.0% |
| `internal/api` | 97.7% |
| `internal/service` | 94.3% |
| `cmd/server` | 0.0% |

`cmd/server` is `main` plus a four-line env helper; testing it means spawning a process.
The arithmetic and validation layers are the ones that carry risk, and both run without
starting an HTTP server.

---

## API

### `POST /api/v1/calculate`

Request:

```json
{ "operation": "add", "operands": [2, 3] }
```

Success — HTTP 200:

```json
{ "result": 5 }
```

Failure — HTTP 400 or 422:

```json
{ "error": { "code": "DIVISION_BY_ZERO", "message": "division by zero" } }
```

Unrecognized extra fields in the request are ignored, not rejected.

### `GET /healthz`

Returns `{"status":"ok"}`. Liveness only.

### Operations

| `operation` | Operands | Meaning |
|---|---|---|
| `add` | 2 | `a + b` |
| `subtract` | 2 | `a - b` |
| `multiply` | 2 | `a * b` |
| `divide` | 2 | `a / b`, divisor must be non-zero |
| `power` | 2 | `a` raised to `b` |
| `sqrt` | 1 | non-negative operand only |
| `percentage` | 2 | `(a / 100) * b` — so `percentage(15, 200)` is `30` |

Operand order is significant for `subtract`, `divide`, `power` and `percentage`.

---

## Examples

Every operation:

```bash
curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"add","operands":[2,3]}'
# {"result":5}

curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"subtract","operands":[10,4]}'
# {"result":6}

curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"multiply","operands":[6,7]}'
# {"result":42}

curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"divide","operands":[10,4]}'
# {"result":2.5}

curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"power","operands":[2,10]}'
# {"result":1024}

curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"sqrt","operands":[9]}'
# {"result":3}

curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"percentage","operands":[15,200]}'
# {"result":30}
```

Results are returned exactly as computed, never rounded:

```bash
curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"add","operands":[0.1,0.2]}'
# {"result":0.30000000000000004}
```

That is not a bug. `0.1 + 0.2` is genuinely `0.30000000000000004` in IEEE 754 binary64.
Rounding it would mean returning a number the service did not compute, so display
formatting is left to the client.

---

## Errors

One envelope for every failure. Branch on `code`; `message` is for humans and may be
reworded without notice.

**400 Bad Request** — the request itself is wrong:

| Code | Cause |
|---|---|
| `MALFORMED_JSON` | Body is not well-formed JSON, or is not an object |
| `MISSING_FIELD` | `operation` or `operands` absent |
| `UNSUPPORTED_OPERATION` | Unknown operation name |
| `INVALID_OPERAND_COUNT` | Operand count does not match the operation's arity |
| `INVALID_OPERAND` | An operand is not a JSON number |
| `OPERAND_OUT_OF_RANGE` | An operand is a number too large to represent |

**422 Unprocessable Entity** — the request is understood, but the calculation has no
representable answer:

| Code | Cause |
|---|---|
| `DIVISION_BY_ZERO` | Divisor is zero |
| `OPERAND_OUT_OF_DOMAIN` | Square root of a negative, or a negative base with a fractional exponent |
| `RESULT_OVERFLOW` | Result is infinite |
| `RESULT_UNDERFLOW` | Result is zero from all-non-zero operands |
| `RESULT_UNDEFINED` | Result is NaN |

The router also returns `404` for an unknown path and `405` for a wrong method. Those are
not calculation failures and carry no JSON body.

### One example per category

```bash
# 422 DIVISION_BY_ZERO
curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"divide","operands":[1,0]}'

# 422 OPERAND_OUT_OF_DOMAIN
curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"sqrt","operands":[-9]}'

# 422 RESULT_OVERFLOW
curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"multiply","operands":[1e308,10]}'

# 422 RESULT_UNDERFLOW
curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"multiply","operands":[1e-200,1e-200]}'

# 400 MALFORMED_JSON
curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":'}

# 400 MISSING_FIELD
curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operands":[1,2]}'

# 400 UNSUPPORTED_OPERATION
curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"tangent","operands":[1,2]}'

# 400 INVALID_OPERAND_COUNT
curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"sqrt","operands":[9,4]}'

# 400 INVALID_OPERAND
curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"add","operands":[1,"abc"]}'

# 400 OPERAND_OUT_OF_RANGE
curl -s -X POST localhost:8080/api/v1/calculate -H 'Content-Type: application/json' \
  -d '{"operation":"add","operands":[1e400,1]}'
```

---

## Project structure

```text
backend/
├── cmd/server/          configuration and startup
└── internal/
    ├── api/             routing, JSON, CORS, HTTP status mapping
    ├── service/         validation, operand parsing, orchestration
    └── calc/            the operation registry, arithmetic, result guards
```

Dependencies point inward: `api` imports `service`, `service` imports `calc`, `calc`
imports only `errors` and `math`. `calc` imports neither `net/http` nor `encoding/json`,
so the arithmetic is testable with no transport in the picture.

---

## Design decisions

### Why Spec-Driven Development?

I used Spec-Driven Development (SDD) for the backend, with
[GitHub Spec Kit](https://github.com/github/spec-kit). The goal was not to add process for
its own sake, but to make the implementation more deliberate and keep engineering decisions
explicit.

- **Specification**: I defined the API behavior, validation rules, edge cases, error
  semantics, and acceptance criteria → `spec.md`.
- **Clarification**: I resolved open questions around numerical behavior and error
  classification → `spec.md`.
- **Research**: I evaluated technical alternatives and implementation constraints →
  `research.md`.
- **Planning**: I defined the architecture and implementation approach → `plan.md`.
- **Tasks**: I broke the implementation into concrete, verifiable tasks → `tasks.md`.
- **Implementation & verification**: I used AI to accelerate implementation, then reviewed
  and verified the result against the requirements and tests.

The complete process is available in
[`specs/001-calculator-rest-api/`](../specs/001-calculator-rest-api/), and the prompts used
during development are available in [`PROMPTS.md`](../PROMPTS.md).

The important distinction is that AI assisted the implementation, but I drove the
engineering process: requirements, design decisions, trade-offs, edge cases, code review,
and verification. This approach let me spend less time on repetitive implementation work and
more time on correctness and testing. The backend took approximately 1.5 hours to implement
using this workflow.

### Architecture & API

- **Layered architecture** (`api → service → calc`): separates HTTP concerns,
  validation/orchestration, and arithmetic logic. The `calc` layer is independent of HTTP,
  making the core behavior easier to test in isolation. I kept the architecture
  intentionally small rather than adding abstractions that were not justified by the scope.
- **API-first**: request/response shapes, operations, and error codes were defined before
  implementation. This forced ambiguous cases to be decided early and gave the frontend a
  stable contract to consume.
- **One endpoint**: all operations use `POST /api/v1/calculate`. This keeps validation and
  error handling consistent and makes adding an operation a registry change rather than
  another route.
- **Operand array**: requests use `operands: [2, 3]` rather than named `a`/`b` fields. This
  makes operation arity explicit and keeps the request structure consistent, including for
  one-operand operations such as `sqrt`.
- **Stable error contract**: all calculation failures use the same JSON envelope with a
  machine-readable `code` and human-readable `message`, allowing clients to handle errors
  without parsing text.
- **Stateless**: each request represents one calculation. No persistence or session state is
  introduced because it is not required by the assignment.

### Numerical behavior

- **`float64` as the numeric type**: calculations use Go's IEEE 754 binary64
  representation. This provides a well-defined range and precision model that is sufficient
  for the scope of the calculator and makes numerical behavior predictable and testable.
- **No arbitrary digit limit**: the API does not impose an application-level maximum number
  of digits. The effective range and precision are determined by the chosen `float64`
  representation rather than an arbitrary input limit.
- **No rounding**: the backend returns the calculated `float64` result without applying
  presentation-level rounding. Formatting is left to the client.
- **Floating-point precision in tests**: comparisons are exact by default. A relative
  tolerance of `1e-9` is used only where the expected value is irrational and therefore not
  exactly representable, such as `sqrt(2)`. Everywhere else the tests require exact
  equality, including `0.1 + 0.2 == 0.30000000000000004`, because a tolerance-based
  comparison could not tell a correct unrounded result from a silently rounded one.
- **`UseNumber()`**: operands are decoded with `json.Decoder.UseNumber()` so the service can
  distinguish malformed or non-numeric operands from numbers that cannot be represented by
  `float64`.
- **Validation before calculation**: validation follows a fixed order — request structure,
  operation, operand count, operand values, then calculation — producing deterministic
  errors.
- **Explicit numerical guards**: division by zero, overflow, underflow, invalid domains, and
  undefined results are handled explicitly.
- **Percentage**: defined as `(a / 100) * b`.
- **Negative base with fractional exponent**: rejected as `OPERAND_OUT_OF_DOMAIN` because
  the calculator operates in the real-number domain.

Detailed reasoning for the numerical decisions is documented in
[`research.md`](../specs/001-calculator-rest-api/research.md).

### Dependencies & testing

- **No third-party dependencies**: Go's standard library provides the routing, JSON handling
  and testing needed for this scope, and CORS is a small middleware written on top of
  `net/http`. This keeps the dependency surface small.
- **Table-driven tests**: tests are organized around the defined behavior and edge cases,
  keeping the mapping between requirements and verification straightforward.
- **Focused coverage**: the arithmetic and validation layers receive the most coverage
  because they contain the core calculation and input-handling risks. The current coverage
  is `100%` for `internal/calc`, `97.7%` for `internal/api`, and `94.3%` for
  `internal/service`.

---

## Assumptions and known limitations

- **REST API for the backend microservice**: I interpreted the requested Go microservice as
  a stateless HTTP REST API using JSON for communication with the frontend. This keeps the
  backend independently deployable and gives the frontend a simple, explicit contract to
  consume.
- **Real numbers only**: the service works in the real-number domain and does not model
  complex results. Operations whose true answer is not a real number are rejected as
  `OPERAND_OUT_OF_DOMAIN` rather than approximated: the square root of a negative operand,
  and a negative base raised to a fractional exponent.
- **One calculation per request**: no batching, and no expression parsing, the service does
  not accept `"2 + 3 * 4"`. A client composes multi-step work from successive requests.
- **No authentication, rate limiting, or persistence**: none were required, and none are
  needed for correctness at this scope.
- **Precision is IEEE 754 binary64 throughout**: roughly 15–17 significant decimal digits.
  Arbitrary-precision and exact decimal arithmetic are out of scope.
- **No request body size limit**: the service reads the request body without a cap. This
  would be the first thing to add before exposing the API publicly.
- **Operands that underflow on input are accepted as zero**: `1e-400` is a well-formed JSON
  number below the representable range; Go parses it to `0` without error, so it is treated
  as `0`. The specification covers operands that *exceed* the range but is silent on this
  case, so it is left as-is and flagged rather than decided unilaterally.
