# Implementation Plan: Calculator REST API

**Branch**: `001-calculator-rest-api` | **Date**: 2026-09-01 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-calculator-rest-api/spec.md`

## Summary

A stateless Go microservice exposing seven arithmetic operations over one HTTP endpoint,
`POST /api/v1/calculate`, with a request body naming the operation and carrying its operands
as an array. Every operand and result is an IEEE 754 binary64 value. Results are returned
exactly as computed, unrounded. Every failure returns the same JSON envelope carrying one of
eleven stable error codes, split across `400` (the request is wrong) and `422` (the
calculation has no representable answer).

Three packages, dependencies pointing inward: `internal/api` handles transport, JSON, CORS
and status mapping; `internal/service` validates and orchestrates; `internal/calc` holds the
operation registry and pure arithmetic and imports nothing beyond `math` and `errors`.
Standard library only — no router, no test framework, no assertion library.

The two decisions that shape the code most: operands are decoded as `[]any` with
`UseNumber()` so that "not a number" and "too large to represent" stay distinguishable
(R4), and a single result-based underflow guard rejects a zero result from all-non-zero
operands — reachable, provably, by only four of the seven operations (R5, FR-024d).

## Technical Context

**Language/Version**: Go 1.22 or later. The version floor is load-bearing: method+pattern
routing in `net/http`'s `ServeMux` is what lets the service avoid a third-party router (R1).
`go` is **not currently on PATH** on this machine and must be installed before Implement.

**Primary Dependencies**: None. Standard library only — `net/http`, `encoding/json`,
`strconv`, `math`, `errors`, `os`, `testing`, `net/http/httptest`. `backend/go.mod` is
expected to have an empty `require` block; a non-empty one is a review failure.

**Storage**: N/A. The service is stateless by requirement (FR-014).

**Testing**: Standard `testing` package, table-driven. `net/http/httptest` for handler
tests. Coverage via `go test -coverprofile`. No `testify`, no mocking framework (R8).

**Target Platform**: Any platform with a Go 1.22+ toolchain. Developed on Windows, must run
unchanged on Linux and macOS. Nothing platform-specific in the code.

**Project Type**: Web service (backend half of a two-part repository; the React frontend is
governed separately and is out of scope for this constitution and this plan).

**Performance Goals**: N/A — explicitly out of scope per the spec's Assumptions. Every
operation is a handful of float instructions; there is no workload to tune. No benchmark,
no load test, no latency target is planned.

**Constraints**: Backend budgeted at roughly two hours within a two-to-four hour full-stack
effort. Correctness ranks above maintainability, testability, and optional features, in that
order. Development proceeds one task at a time with a developer review and manual commit
between tasks; the AI agent never commits (Constitution VI and VII).

**Scale/Scope**: Two endpoints, seven operations, eleven error codes, three packages.
Estimated 400-600 lines of Go including tests. Single developer, no concurrency concerns
beyond what `net/http` already handles — the service holds no shared mutable state.

## Constitution Check

*GATE: evaluated before Phase 0, re-evaluated after Phase 1 design. Both passes clean.*

| # | Principle | Verdict | Evidence |
|---|---|---|---|
| I | API-First | **PASS** | Contract fixed before implementation in [contracts/openapi.yaml](contracts/openapi.yaml) and [contracts/error-codes.md](contracts/error-codes.md). No UI concern anywhere in the backend; the service assumes nothing about its client. Stateless per FR-014. |
| II | Layered Architecture | **PASS** | Exactly three packages — `api`, `service`, `calc` — with the import rules tabulated in [data-model.md](data-model.md). `calc` imports neither `net/http` nor `encoding/json`. No fourth layer. |
| III | Testability | **PASS** | `calc` and `service` are testable with no server running (FR-028). Every error code gets at least one test (FR-030). Coverage report is a planned task. Tests must pass before each task is presented for review. |
| IV | Simplicity and Idiomatic Go | **PASS** | Zero third-party dependencies. Each avoided dependency is recorded with its rationale in [research.md](research.md) — `chi`/`gin` (R1), `testify` (R8), `rs/cors` (R9). |
| V | Avoid Overengineering | **PASS** | No interfaces, no DI, no codegen, no persistence, no caching, no CI/CD, no observability stack. Containerization explicitly deferred (R10). The one configuration concession is two `os.Getenv` calls with defaults (R9) — values that genuinely vary between dev and any other environment, not a configuration system. |
| VI | Incremental Development | **PASS** | `tasks.md` will be ordered so each task ends with passing tests and a review stop. Enforced at Implement time, not here. |
| VII | Human-Controlled Version Control | **PASS** | No task in this plan creates a commit, branch, or tag. The developer commits. |

**Complexity Tracking**: not required — no violations to justify.

One item to watch during Implement, recorded here rather than as a violation: `service`
receives operands as `[]any` containing `json.Number`, which is a type from
`encoding/json`. This is a deliberate, minimal transport leak into the validation layer,
and it is what makes the FR-021 / FR-022 split possible without parsing Go's internal error
strings (R4). The alternative — classifying tokens in `api` and passing `[]float64` inward —
would move validation into the transport layer, a worse violation of Constitution II than
one type name. Flagged for the reviewer.

## Project Structure

### Documentation (this feature)

```text
specs/001-calculator-rest-api/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0 output — 10 decisions with alternatives
├── data-model.md        # Phase 1 output — entities, validation pipeline, layer ownership
├── quickstart.md        # Phase 1 output — run instructions and 18 validation scenarios
├── contracts/
│   ├── openapi.yaml     # Phase 1 output — HTTP contract
│   └── error-codes.md   # Phase 1 output — the eleven error identifiers
├── checklists/
│   └── requirements.md  # Spec quality checklist, 16/16
└── tasks.md             # Phase 2 output — created by /speckit-tasks, NOT by this command
```

### Source Code (repository root)

```text
backend/
├── go.mod                          # module calculator/backend — no third-party requires
├── cmd/
│   └── server/
│       └── main.go                 # config from env, build router, ListenAndServe
└── internal/
    ├── calc/
    │   ├── operation.go            # Operation struct, registry, arity
    │   ├── errors.go               # domain and result sentinel errors
    │   ├── operations.go           # add/subtract/multiply/divide/power/sqrt/percentage
    │   ├── guard.go                # non-finite result check, underflow-to-zero check
    │   └── *_test.go               # table-driven, one row per spec edge case
    ├── service/
    │   ├── calculate.go            # steps 1-6: presence, lookup, arity, operand parsing
    │   ├── errors.go               # service error type carrying a stable code
    │   └── calculate_test.go       # validation and classification, no HTTP
    └── api/
        ├── router.go               # ServeMux, route registration, CORS middleware
        ├── handler.go              # decode, delegate, encode result or error envelope
        ├── status.go               # code -> HTTP status table
        └── handler_test.go         # httptest round trips, one per error code

frontend/                           # Out of scope for this plan and this constitution
```

**Structure Decision**: Two top-level directories already exist in the repository,
`backend/` and `frontend/`, so the web-application layout applies. This plan governs
`backend/` only. Inside it, `cmd/` plus `internal/` is the standard Go layout; `internal/`
prevents the packages being imported by anything outside the module, which is correct for a
service. The three subpackages of `internal/` map one-to-one onto the three layers
Constitution II mandates, so the layering is visible in the directory tree and a violation
is visible in an import statement.

## Phase 0 — Research

Complete. See [research.md](research.md). Ten decisions recorded, each with rationale and
rejected alternatives:

| ID | Decision |
|---|---|
| R1 | Go 1.22+, stdlib `ServeMux` with method patterns |
| R2 | One endpoint, `POST /api/v1/calculate`, plus `GET /healthz` |
| R3 | `{"operation": "...", "operands": [...]}` — array, not named fields |
| R4 | `[]any` + `UseNumber()` to separate malformed / not-a-number / out-of-range |
| R5 | Result-based underflow guard; provably reachable only by multiply, divide, power, percentage |
| R6 | Eleven error codes; 400 for bad requests, 422 for unanswerable calculations |
| R7 | Negative base with non-integer exponent pre-checked as a domain error |
| R8 | Table-driven stdlib tests, two-tier comparison per FR-031a |
| R9 | Minimal CORS middleware, `CORS_ORIGIN` and `PORT` env vars with defaults |
| R10 | Containerization deferred — not requested, forbidden by Constitution V until it is |

No NEEDS CLARIFICATION markers remain.

## Phase 1 — Design & Contracts

Complete. Artifacts:

- **[data-model.md](data-model.md)** — the `Operation` registry (name, arity, domain rule,
  whether underflow-checked), the request/result/error shapes, the ten-step validation
  pipeline with the failing identifier for each step, and the import rules per layer.
- **[contracts/openapi.yaml](contracts/openapi.yaml)** — both endpoints, request and
  response schemas, worked examples including `0.1 + 0.2` returning
  `0.30000000000000004`, and the error code enum.
- **[contracts/error-codes.md](contracts/error-codes.md)** — the eleven identifiers with
  their trigger conditions and originating requirements, plus an explicit list of cases that
  are *not* errors, which is where the FR-024b legitimate-zero rules live.
- **[quickstart.md](quickstart.md)** — prerequisites, run and test commands, and 18
  numbered validation scenarios covering every success path and every error code.

### Requirement coverage

Every functional requirement has a home in the design:

| Requirements | Where satisfied |
|---|---|
| FR-001 to FR-008, FR-007a | `calc` registry — one entry per operation, arity as data |
| FR-009 to FR-017, FR-017a | `api` handler and envelope; `openapi.yaml` |
| FR-012a, FR-024c | `contracts/error-codes.md` — eleven codes, none shared |
| FR-018 to FR-023, FR-022a | `service` validation pipeline, steps 1-6 |
| FR-024, FR-024a, FR-024b, FR-024d | `calc/guard.go` — non-finite check, then underflow check; guard attached to the four operations FR-024d permits |
| FR-025, FR-026, R7 | Per-operation domain checks in `calc/operations.go` |
| FR-027, FR-036 | No length, magnitude or precision caps anywhere in validation |
| FR-028 to FR-031, FR-031a to FR-031c | Test layout; `calc` and `service` tests need no server; subnormal-boundary rows in the subtract table per FR-031c |
| FR-032 to FR-035 | `float64` throughout; operand classification in `service` |

### Post-design Constitution re-check

Re-evaluated against the finished design: **PASS**, unchanged from the pre-Phase-0 gate.
The design added no dependency, no fourth layer, no abstraction, and no configuration
surface beyond the two documented environment variables.

## Risks and watch items

| Risk | Mitigation |
|---|---|
| Go 1.22+ not installed — blocks Implement entirely | Install before the first task. Verified by `go version` in quickstart. |
| The FR-024a underflow guard is unusual and easy to over-apply, breaking legitimate zeros | Scenario 15 in quickstart exists solely to catch this; `CheckUnderflow` is data on the registry entry, not a scattered condition; FR-031c adds subnormal-boundary tests that fail loudly if the guard is widened |
| A future maintainer "fixes" the guard by extending it to addition and subtraction | FR-024d records the proof that they cannot underflow to zero, and states the widening prohibition inline rather than leaving the operation list looking arbitrary |
| `0.1 + 0.2` silently rounded by a JSON encoder or a helpful `fmt` verb | FR-031b test plus quickstart scenario 4; encode with `encoding/json` defaults, never `strconv.FormatFloat` with a fixed precision |
| Two-hour budget consumed by the eleven-code error taxonomy | The taxonomy is data (a map) plus one status table; the cost is in tests, which are the deliverable anyway |
| Scope drift into the frontend | Constitution's Scope and Constraints puts the frontend out of scope; this plan touches `backend/` only |
