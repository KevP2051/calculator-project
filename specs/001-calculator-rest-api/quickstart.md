# Quickstart: Calculator REST API

**Feature**: 001-calculator-rest-api | **Date**: 2026-09-01

How to run the backend and prove it satisfies the spec. This is a validation guide, not an
implementation guide — the code belongs to `tasks.md` and the Implement phase.

---

## Prerequisites

- **Go 1.22 or later.** Required for `net/http`'s method+pattern routing (R1).

Verify:

```bash
go version
```

No database, no message broker, no container runtime. The service has no dependencies
beyond the Go standard library.

---

## Run the service

From `backend/`:

```bash
go run ./cmd/server
```

Listens on `:8080` by default. Override with `PORT`. Allowed CORS origin defaults to
`http://localhost:5173`; override with `CORS_ORIGIN`.

Confirm it is up:

```bash
curl -s http://localhost:8080/healthz
```

Expected: `{"status":"ok"}`

---

## Run the tests

From `backend/`:

```bash
go test ./... -cover
```

Every package should pass. Core arithmetic tests run without starting an HTTP server
(FR-028, SC-007).

Coverage report:

```bash
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html
```

Target: at least 80% on `internal/calc` and `internal/service`.

---

## Validation scenarios

Each scenario below maps to a spec requirement. Run them against a live server. Together
they cover every success path and every error identifier in
[contracts/error-codes.md](contracts/error-codes.md).

### 1. The four basic operations (User Story 1, FR-001 to FR-004)

```bash
curl -s -X POST http://localhost:8080/api/v1/calculate -H 'Content-Type: application/json' -d '{"operation":"add","operands":[2,3]}'
```

Expected: `{"result":5}`. Repeat with `subtract` `[10,4]` for `6`, `multiply` `[6,7]` for
`42`, `divide` `[10,4]` for `2.5`.

### 2. Statelessness (FR-014, FR-017, SC-008)

Send the same request twice. Both responses must be byte-identical.

### 3. Extended operations (User Story 3, FR-005 to FR-007)

```bash
curl -s -X POST http://localhost:8080/api/v1/calculate -H 'Content-Type: application/json' -d '{"operation":"percentage","operands":[15,200]}'
```

Expected: `{"result":30}`. Also check `power` `[2,10]` for `1024` and `sqrt` `[9]` for `3`.

### 4. Results are unrounded (FR-017a, FR-031b, SC-007a)

```bash
curl -s -X POST http://localhost:8080/api/v1/calculate -H 'Content-Type: application/json' -d '{"operation":"add","operands":[0.1,0.2]}'
```

Expected: `{"result":0.30000000000000004}` — **not** `{"result":0.3}`. This is the single
most important check in the file: it is the only one that can detect a service that silently
rounds.

### 5. Division by zero (FR-025)

```bash
curl -s -i -X POST http://localhost:8080/api/v1/calculate -H 'Content-Type: application/json' -d '{"operation":"divide","operands":[1,0]}'
```

Expected: `422`, code `DIVISION_BY_ZERO`, no `result` field anywhere in the body.

### 6. Square root of a negative number (FR-026)

`{"operation":"sqrt","operands":[-9]}` — expected `422`, code `OPERAND_OUT_OF_DOMAIN`.

### 7. Malformed JSON (FR-019)

Send `{"operation":` — expected `400`, code `MALFORMED_JSON`.

### 8. Missing field (FR-020)

`{"operands":[1,2]}` — expected `400`, code `MISSING_FIELD`.

### 9. Unsupported operation (FR-008)

`{"operation":"tangent","operands":[1,2]}` — expected `400`, code
`UNSUPPORTED_OPERATION`. The service must not fall back to a default operation.

### 10. Wrong operand count (FR-023, FR-007a)

`{"operation":"sqrt","operands":[9,4]}` — expected `400`, code `INVALID_OPERAND_COUNT`.

### 11. Operand that is not a number (FR-021, FR-035)

`{"operation":"add","operands":[1,"abc"]}` — expected `400`, code `INVALID_OPERAND`.
Repeat with `"NaN"`, `"Infinity"`, `true`, and `null`; all four give the same code.

### 12. Operand out of range (FR-022, FR-022a)

`{"operation":"add","operands":[1e400,1]}` — expected `400`, code `OPERAND_OUT_OF_RANGE`.
Note this is a **different** code from scenario 11: that one was not a number, this one is a
number too large to represent.

### 13. Result overflow (FR-024)

`{"operation":"multiply","operands":[1e308,10]}` — expected `422`, code `RESULT_OVERFLOW`.
The response must not contain `Infinity` (FR-016, SC-003).

### 14. Result underflow (FR-024a)

`{"operation":"multiply","operands":[1e-200,1e-200]}` — expected `422`, code
`RESULT_UNDERFLOW`. The true answer is `1e-400`, too small to represent; the service must
not report `0`.

### 15. Legitimate zeros are NOT underflow (FR-024b)

All of these must return `200` with `{"result":0}`:

- `{"operation":"subtract","operands":[5,5]}` — exact cancellation
- `{"operation":"multiply","operands":[0,5]}` — zero operand
- `{"operation":"percentage","operands":[5,0]}` — zero operand
- `{"operation":"sqrt","operands":[0]}` — square root cannot underflow

If any of these returns `RESULT_UNDERFLOW`, the underflow guard is too broad.

### 15b. Subtraction at the subnormal boundary (FR-024d, FR-031c)

Binary64 addition and subtraction cannot underflow to zero, so both of these are successes:

- `{"operation":"subtract","operands":[5e-324,5e-324]}` — expected `{"result":0}`, exact
  cancellation at the smallest positive subnormal
- `{"operation":"subtract","operands":[1.5e-323,1e-323]}` — expected a non-zero subnormal
  result, roughly `5e-324`

A `RESULT_UNDERFLOW` from either means the guard has been widened past what FR-024d
permits.

### 16. Extra fields are ignored (Edge Cases)

`{"operation":"add","operands":[2,3],"precision":4,"nonsense":true}` — expected `200`,
`{"result":5}`. Extra fields are not an error, and `precision` in particular must have no
effect, since the service never rounds.

### 17. Router-level responses

```bash
curl -s -i -X GET http://localhost:8080/api/v1/calculate
```

Expected: `405`. And any unknown path gives `404`.

---

## Definition of done for this feature

- All 18 scenarios above behave as described.
- `go test ./... -cover` passes with no failures.
- No third-party module appears in `backend/go.mod` beyond the standard library.
- No response ever contains `Infinity`, `-Infinity`, or `NaN`.
- `internal/calc` imports neither `net/http` nor `encoding/json`.
