---

description: "Task list for Calculator REST API"
---

# Tasks: Calculator REST API

**Input**: Design documents from `/specs/001-calculator-rest-api/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/](contracts/), [quickstart.md](quickstart.md)

**Tests**: REQUIRED. The specification mandates them (FR-028 through FR-031c) and
Constitution III makes passing tests a precondition for presenting any task as complete.

**Organization**: Grouped by user story. Each phase is a reviewable increment.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Touches different files from its neighbours and has no dependency on incomplete
  work — parallelizable in principle.
- **[Story]**: US1, US2, US3 — maps to the user stories in spec.md.
- Every task names its exact file paths.

## Path Conventions

Web application layout. This feature governs `backend/` only; `frontend/` is out of scope
per the constitution's Scope and Constraints.

## Execution rules for this feature

These override the generic guidance in the template:

1. **One task per turn.** Constitution VI: after each task the agent stops, reports what
   changed, and waits. No batching, no working ahead.
2. **The developer commits.** Constitution VII: the agent never runs `git commit`, `git
   push`, or any branch operation.
3. **Tests ship with their code, in the same task.** Constitution III requires tests to pass
   before a task is presented as complete, so a task that leaves a deliberately failing test
   cannot be presented at all. Red-green is therefore practiced *inside* a task, not split
   across two of them.
4. **[P] is advisory.** With one developer and a review gate between every task, execution is
   sequential in practice. The markers record which tasks have no ordering constraint, so a
   reviewer can reorder if something is blocked.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: An empty but valid Go module with the directory skeleton from plan.md.

- [X] T001 Initialize the Go module and package skeleton: create `backend/go.mod`
      (`module calculator/backend`, `go 1.22`, empty `require` block) and the empty
      directories `backend/cmd/server/`, `backend/internal/calc/`,
      `backend/internal/service/`, `backend/internal/api/`. Verify with
      `cd backend && go build ./...`.
- [X] T002 [P] Add Go build artifacts to `.gitignore` at the repository root: compiled
      binaries (`backend/server`, `backend/server.exe`). Do **not** ignore
      `backend/coverage.out` or `backend/coverage.html` — the assignment lists a coverage
      report among its deliverables, so T017 commits both.

**Checkpoint**: `go build ./...` succeeds from `backend/`. Nothing else works yet.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The error vocabulary, the result guards, and a booting server. Every user story
depends on all of it.

**CRITICAL**: No user story work begins until this phase is complete.

- [X] T003 Define the core failure sentinels in `backend/internal/calc/errors.go`:
      `ErrDivisionByZero`, `ErrOutOfDomain`, `ErrOverflow`, `ErrUnderflow`, `ErrUndefined`.
      Plain `errors.New` sentinels, comparable with `errors.Is`. This file must import
      nothing beyond `errors` (Constitution II).
- [X] T004 [P] Define the service error type and the eleven stable codes in
      `backend/internal/service/errors.go`, exactly as catalogued in
      [contracts/error-codes.md](contracts/error-codes.md): `MALFORMED_JSON`,
      `MISSING_FIELD`, `UNSUPPORTED_OPERATION`, `INVALID_OPERAND_COUNT`, `INVALID_OPERAND`,
      `OPERAND_OUT_OF_RANGE`, `DIVISION_BY_ZERO`, `OPERAND_OUT_OF_DOMAIN`,
      `RESULT_OVERFLOW`, `RESULT_UNDERFLOW`, `RESULT_UNDEFINED`. Include the mapping from
      each `calc` sentinel to its code. Must not import `net/http` (FR-012a).
- [X] T005 [P] Implement the code-to-HTTP-status table in `backend/internal/api/status.go`
      (400 for the six request-fault codes, 422 for the five calculation-fault codes) with a
      table test in `backend/internal/api/status_test.go` asserting every one of the eleven
      codes maps to a non-zero status. A code with no mapping must fail the test, not
      default silently (R6).
- [X] T006 Define the `Operation` struct and the empty registry in
      `backend/internal/calc/operation.go`: fields `Name`, `Arity`, `Apply`,
      `CheckUnderflow`, plus a `Lookup(name string) (Operation, bool)`. Registry starts
      empty; operations are registered by their own tasks. Per
      [data-model.md](data-model.md).
- [X] T007 Implement the result guards in `backend/internal/calc/guard.go` with tests in
      `backend/internal/calc/guard_test.go`: a non-finite check mapping `+Inf`/`-Inf` to
      `ErrOverflow` and `NaN` to `ErrUndefined` (FR-024), and an underflow check returning
      `ErrUnderflow` when the result is zero and every operand is non-zero (FR-024a). Tests
      must cover: a zero result with a zero operand is NOT underflow, and `-0.0` is treated
      as zero. Add a comment on the underflow guard citing FR-024d — it must never be
      attached to `add`, `subtract`, or `sqrt`.
- [X] T008 Build the server bootstrap: `backend/internal/api/router.go` (a `ServeMux` with
      `GET /healthz` returning `{"status":"ok"}`, plus CORS middleware reading `CORS_ORIGIN`
      with default `http://localhost:5173`, handling `OPTIONS` preflight) and
      `backend/cmd/server/main.go` (reads `PORT`, default `8080`, then `ListenAndServe`).
      Test `/healthz` and the CORS headers in `backend/internal/api/router_test.go` using
      `httptest`. Per R9.

**Checkpoint**: `go run ./cmd/server` starts, `curl localhost:8080/healthz` returns
`{"status":"ok"}`, `go test ./...` passes. No arithmetic exists yet.

---

## Phase 3: User Story 1 - Perform a basic arithmetic calculation (Priority: P1) 🎯 MVP

**Goal**: A client can POST addition, subtraction, multiplication or division with valid
operands and receive the correct result as JSON.

**Independent Test**: Run quickstart scenarios 1, 2, 4 and 15. Four operations return correct
values, repeated requests are identical, and `0.1 + 0.2` returns `0.30000000000000004`.

**Scope note**: T010 implements the whole validation pipeline, including the branches that
return US2's error codes, because they are branches of one function — splitting them across
phases would mean writing the function twice. US1 tests its happy path; US2 tests those
branches exhaustively and is where they are verified.

- [X] T009 [US1] Implement `add`, `subtract`, `multiply`, `divide` in
      `backend/internal/calc/operations.go` and register them in the T006 registry
      (`CheckUnderflow` true for `multiply` and `divide`, false for `add` and `subtract`,
      per FR-024d). `divide` returns `ErrDivisionByZero` when the divisor is zero, before
      computing (FR-025). Table-driven tests in
      `backend/internal/calc/operations_test.go` covering: correct results for each
      operation; operand order for subtract and divide; exact equality on
      `0.1 + 0.2 == 0.30000000000000004` (FR-031b); `5 - 5` and `5 + -5` returning `0` as a
      success; `0 * 5` and `0 / 5` returning `0` as a success; the two subnormal-boundary
      subtraction rows required by FR-031c (`5e-324 - 5e-324` returning `0`, and
      `1.5e-323 - 1e-323` returning a non-zero subnormal); `1e308 * 10` returning
      `ErrOverflow`; `1e-200 * 1e-200` returning `ErrUnderflow`. Comparison follows FR-031a:
      exact equality throughout this task, since every expected value here is exactly
      representable.
- [X] T010 [US1] Implement the validation pipeline in
      `backend/internal/service/calculate.go`: accept the decoded request shape
      (`Operation string`, `Operands []any` with `json.Number` elements), then run steps 2-6
      of [data-model.md](data-model.md) — field presence, registry lookup, arity check,
      operand classification via `strconv.ParseFloat` with `strconv.ErrRange` distinguishing
      `OPERAND_OUT_OF_RANGE` from `INVALID_OPERAND` (R4) — dispatch to the operation, then
      apply the T007 guards. Tests in `backend/internal/service/calculate_test.go` covering
      the happy path for all four operations and confirming no HTTP import is needed to run
      them (FR-028).
- [X] T011 [US1] Implement the calculate handler in `backend/internal/api/handler.go`:
      decode the body with `json.Decoder` and `UseNumber()`, delegate to the T010 service,
      and encode `{"result": <number>}` on success using `encoding/json` defaults — never
      `strconv.FormatFloat` with a fixed precision, which would violate FR-017a. Register
      `POST /api/v1/calculate` in `backend/internal/api/router.go`. Round-trip tests in
      `backend/internal/api/handler_test.go` using `httptest`, asserting the four operations
      and that the `0.1 + 0.2` response body literally contains `0.30000000000000004`.

**Checkpoint**: A working four-function calculator over HTTP. Quickstart scenarios 1, 2, 4,
15, 15b and 17 pass. This is the MVP — stop here and validate before continuing.

---

## Phase 4: User Story 2 - Receive predictable, machine-readable errors (Priority: P2)

**Goal**: Every failure returns the shared envelope with a distinct stable code and the right
status, and never a numeric result.

**Independent Test**: Run quickstart scenarios 5 through 14. Each returns its own code, all
share one envelope shape, and no error body contains a `result` field.

- [X] T012 [US2] Complete and verify the validation error branches in
      `backend/internal/service/calculate.go`, adding a test case per code in
      `backend/internal/service/calculate_test.go`: `MISSING_FIELD` (absent `operation`,
      absent `operands`), `UNSUPPORTED_OPERATION` (an unknown name, asserting no fallback to
      a default operation per FR-008), `INVALID_OPERAND_COUNT` (two operands for `sqrt`, one
      for `add`), `INVALID_OPERAND` (string, boolean, null, `"NaN"`, `"Infinity"` — all one
      code per FR-035), and `OPERAND_OUT_OF_RANGE` (`1e400`, `-1e400`). Assert explicitly
      that `INVALID_OPERAND` and `OPERAND_OUT_OF_RANGE` are different codes (FR-022a).
- [X] T013 [US2] Implement the error envelope in `backend/internal/api/handler.go`:
      `{"error":{"code":"...","message":"..."}}`, status from the T005 table, and
      `MALFORMED_JSON` for a body that fails to decode or is not a JSON object (FR-019).
      Confirm extra unrecognized fields are ignored rather than rejected. Tests in
      `backend/internal/api/handler_test.go`: one round trip per code asserting status, code,
      and the absence of any `result` key in the body (FR-015); plus a test asserting all
      eleven error responses share the same top-level shape (FR-011, SC-006).
- [X] T014 [US2] Add the calculation-failure round trips to
      `backend/internal/api/handler_test.go`: `DIVISION_BY_ZERO` (422), `RESULT_OVERFLOW`
      (`1e308 * 10`), `RESULT_UNDERFLOW` (`1e-200 * 1e-200`), and a test asserting no
      response body ever contains the tokens `Infinity`, `-Infinity`, or `NaN` (FR-016,
      SC-003). Also assert `subtract` at the subnormal boundary returns `200`, guarding
      FR-024d against a future widening of the underflow check (SC-007b).

**Checkpoint**: Every error path in the contract is implemented and tested. Quickstart
scenarios 5-14 pass.

---

## Phase 5: User Story 3 - Perform extended operations (Priority: P3)

**Goal**: Exponentiation, square root and percentage work, with their domain errors reported
as errors rather than as `NaN` or infinity.

**Independent Test**: Run quickstart scenarios 3 and 6. The three operations return correct
values, and `sqrt(-9)` plus a negative base with a fractional exponent both return
`OPERAND_OUT_OF_DOMAIN`.

- [ ] T015 [US3] Implement `power` in `backend/internal/calc/operations.go` and register it
      (`Arity` 2, `CheckUnderflow` true). Pre-check `base < 0 && exp != math.Trunc(exp)` and
      return `ErrOutOfDomain` *before* calling `math.Pow`, so the case reports as a domain
      error rather than `RESULT_UNDEFINED` (R7). Tests in
      `backend/internal/calc/operations_test.go`: `2^10 == 1024`; `0^0 == 1`; `0^-1`
      returning `ErrOverflow`; `(-8)^(1/3)` returning `ErrOutOfDomain`; `(-2)^3 == -8`
      (integer exponents on a negative base remain valid); `2^-100000` returning
      `ErrUnderflow`.
- [ ] T016 [US3] Implement `sqrt` (`Arity` 1, `CheckUnderflow` false) and `percentage`
      (`Arity` 2, `CheckUnderflow` true, computing `(a / 100) * b`) in
      `backend/internal/calc/operations.go` and register both. `sqrt` returns
      `ErrOutOfDomain` for a negative operand (FR-026). Tests in
      `backend/internal/calc/operations_test.go`: `sqrt(9) == 3` and `sqrt(0) == 0` exactly;
      `sqrt(2)` within a relative tolerance of `1e-9` (FR-031a — the only place in this
      feature where the tolerance tier applies); `sqrt(-9)` returning `ErrOutOfDomain`;
      `percentage(15, 200) == 30` exactly (FR-007); `percentage(5, 0)` and
      `percentage(0, 5)` returning `0` as successes (FR-024b).

**Checkpoint**: All seven operations work. Every quickstart scenario should now pass.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T017 Generate the coverage report from `backend/`:
      `go test ./... -coverprofile=coverage.out` then
      `go tool cover -html=coverage.out -o coverage.html`. Confirm at least 80% on
      `internal/calc` and `internal/service`. If either falls short, add the missing cases
      rather than lowering the bar. Commit both files — the assignment asks for a coverage
      report as a deliverable, which is why T002 deliberately leaves them out of
      `.gitignore`.
- [ ] T018 Verify constitutional compliance as a reviewable checklist, recording the result
      in the task report: `backend/go.mod` has an empty `require` block (Constitution IV);
      `backend/internal/calc/` imports neither `net/http` nor `encoding/json`
      (Constitution II) — check with
      `grep -rE '"net/http"|"encoding/json"' backend/internal/calc/`; no interface with a
      single implementation and no configuration beyond `PORT` and `CORS_ORIGIN`
      (Constitution V).
- [ ] T019 Run all 18 validation scenarios in
      `specs/001-calculator-rest-api/quickstart.md` against a live server started with
      `go run ./cmd/server` from `backend/`, and record the outcome of each. Any deviation is a defect to fix before this
      task is presented as complete, not a note to file.
- [ ] T020 [P] Write `backend/README.md`: setup and prerequisites, how to run the server and
      the tests, `curl` examples for every operation **and** for each error category, the
      environment variables, and a Design Decisions section drawn from
      [research.md](research.md) — the single-endpoint choice (R2), the operand array (R3),
      standard-library-only (R1, R8, R9), the operand classification strategy (R4), and the
      underflow rule with its FR-024d derivation (R5).

---

## Dependencies & Execution Order

### Phase dependencies

- **Setup (Phase 1)**: no dependencies.
- **Foundational (Phase 2)**: needs Phase 1. Blocks every user story.
- **US1 (Phase 3)**: needs Phase 2.
- **US2 (Phase 4)**: needs Phase 3 — it completes and tests the error branches of the
  pipeline T010 introduces.
- **US3 (Phase 5)**: needs Phase 2 only. Genuinely independent of US1 and US2; it adds
  registry entries and touches no shared logic.
- **Polish (Phase 6)**: needs every story you intend to ship.

### Task-level dependencies

```text
T001 → T002
T001 → T003 → T004 → T005
T003 → T006 → T007
T005, T006 → T008
T006, T007 → T009 → T010 → T011        (US1)
T010 → T012 → T013 → T014              (US2)
T006, T007 → T015, T016                (US3, parallel with US1/US2)
T011, T014, T016 → T017 → T018 → T019 → T020
```

### Parallel opportunities

- T004 and T005 are independent of each other once T003 exists.
- T015 and T016 could run alongside Phase 3 and 4 with a second developer, since they only
  add registry entries.
- T020 can be drafted at any point after Phase 5.

In practice all of this is sequential: one developer, one review gate per task.

---

## Implementation Strategy

### MVP first

1. Phase 1 (T001-T002) — module exists.
2. Phase 2 (T003-T008) — server boots, error vocabulary and guards in place.
3. Phase 3 (T009-T011) — four-function calculator over HTTP.
4. **Stop and validate**: quickstart scenarios 1, 2, 4, 15, 15b, 17.

That is a demonstrable deliverable in eleven tasks. Everything after it is additive.

### Incremental delivery

Phase 4 makes the service integrable (a client can branch on failures). Phase 5 completes
the assignment's optional operations. Phase 6 produces the artifacts a reviewer reads first:
coverage report and README.

### If the budget runs short

The backend is budgeted at roughly two hours and this list has twenty review gates, which is
the main risk to that budget. If time is tight, these merges preserve reviewability:

- T004 + T005 into one error-vocabulary task.
- T012 + T013 + T014 into one "all error paths" task.
- T015 + T016 into one extended-operations task.

That takes twenty tasks to sixteen. Do not merge across a phase checkpoint, and do not drop
T017 or T020 — the coverage report and README are explicit assignment deliverables, not
polish.

---

## Notes

- Every task ends with `go test ./...` passing, then a stop for developer review and a manual
  commit. The agent commits nothing.
- Ambiguity found mid-task that would change the API contract or the architecture goes to the
  developer rather than being resolved by assumption (Constitution, Development Workflow).
- Requirement IDs in task descriptions are load-bearing: they are how a reviewer checks that
  the increment did what the spec asked.
