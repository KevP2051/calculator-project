# Phase 1 Data Model: Calculator REST API

**Feature**: 001-calculator-rest-api | **Date**: 2026-09-01

The service is stateless and stores nothing. "Data model" here means the in-memory shapes
that cross layer boundaries, plus the operation registry that drives validation.

---

## Numeric type

Every operand and every result is an IEEE 754 binary64 value — Go's `float64` (FR-032). No
other numeric type appears in any signature.

---

## Entities

### Operation

The registry entry that drives validation and dispatch. Fixed at compile time; there is no
runtime registration.

| Field | Type | Notes |
|---|---|---|
| `Name` | `string` | Wire identifier, lowercase. The map key. |
| `Arity` | `int` | `1` for square root, `2` for all others (FR-007a). |
| `Apply` | `func([]float64) (float64, error)` | Pure computation plus its own domain checks. |
| `CheckUnderflow` | `bool` | Whether FR-024a's zero-result check applies. Derived from FR-024d, not chosen. |

Registry contents:

| Name | Arity | Domain restriction | Underflow-checked |
|---|---|---|---|
| `add` | 2 | none | no |
| `subtract` | 2 | none | no |
| `multiply` | 2 | none | yes |
| `divide` | 2 | divisor must be non-zero (FR-025) | yes |
| `power` | 2 | negative base requires an integer exponent (R7) | yes |
| `sqrt` | 1 | operand must be non-negative (FR-026) | no |
| `percentage` | 2 | none | yes |

`percentage(a, b) = (a / 100) * b` (FR-007).

Only four operations are underflow-checked, and the list is derived rather than chosen
(FR-024d, R5). Addition and subtraction cannot underflow to zero in binary64 — every finite
value is an integer multiple of `2^-1074`, so a non-zero exact sum or difference is at least
that large and stays representable — which makes any zero they produce exact cancellation.
Square root is excluded because `sqrt(x) == 0` only when `x == 0`. Setting `CheckUnderflow`
on `add` or `subtract` would not add detection; it would reject `5 - 5` and break FR-024b.

### Calculation Request

What arrives from the client, after JSON decoding and before validation.

| Field | Type | Notes |
|---|---|---|
| `Operation` | `string` | Empty string means the field was absent. |
| `Operands` | `[]any` | Decoded with `UseNumber`, so numbers arrive as `json.Number`. `nil` means the field was absent; an empty non-nil slice means `[]` was sent. |

Unrecognized extra fields are discarded by the decoder and are not an error (spec Edge
Cases).

### Validated Operands

The output of the validation step: `[]float64`, every element finite, length equal to the
selected operation's arity. This is the only shape the core layer accepts, so the core
never revalidates.

### Calculation Result

A single finite `float64`. Serialized as a bare JSON number, unrounded and unformatted
(FR-017a).

### Calculation Error

| Field | Type | Notes |
|---|---|---|
| `Code` | `string` | One of the eleven identifiers in [contracts/error-codes.md](contracts/error-codes.md). Stable; part of the contract. |
| `Message` | `string` | Human-readable, for developers. Never parsed by clients (FR-012). |
| `Status` | `int` | HTTP status, assigned by the mapping table, not by the core layer. |

Never carries a numeric result (FR-015).

---

## Validation pipeline

Order matters: FR-018 requires the whole request to be validated before any calculation, and
each step's failure has a distinct identifier.

| # | Step | Failure identifier |
|---|---|---|
| 1 | Body decodes as a JSON object | `MALFORMED_JSON` |
| 2 | `operation` present and non-empty; `operands` present | `MISSING_FIELD` |
| 3 | `operation` names a registry entry | `UNSUPPORTED_OPERATION` |
| 4 | `len(operands)` equals the entry's arity | `INVALID_OPERAND_COUNT` |
| 5 | Each element is a `json.Number` | `INVALID_OPERAND` |
| 6 | Each parses to a finite `float64` | `OPERAND_OUT_OF_RANGE` |
| 7 | Operands satisfy the operation's domain | `DIVISION_BY_ZERO`, `OPERAND_OUT_OF_DOMAIN` |
| 8 | — compute — | |
| 9 | Result is finite | `RESULT_OVERFLOW`, `RESULT_UNDEFINED` |
| 10 | Result is not an underflow to zero | `RESULT_UNDERFLOW` |

Steps 1-6 run in the service layer on transport-shaped input. Steps 7-10 belong to the core
layer and are expressed in terms of `float64` only.

Step 6 detail: `strconv.ParseFloat` returning `strconv.ErrRange` yields
`OPERAND_OUT_OF_RANGE`; any other parse failure, or a resulting value that is `Inf` or
`NaN`, yields `INVALID_OPERAND` (R4, FR-035).

Step 9 detail: `math.IsInf(r, 0)` yields `RESULT_OVERFLOW`; `math.IsNaN(r)` yields
`RESULT_UNDEFINED` (FR-024).

Step 10 detail: fires when the result is zero while every operand is non-zero (FR-024a).
Reachable only for the four underflow-checked operations, per FR-024d. Otherwise a zero
result is returned as a success (FR-024b).

---

## Layer ownership

| Layer | Package | Owns | Must not import |
|---|---|---|---|
| API / infrastructure | `internal/api` | routing, JSON decode/encode, CORS, status mapping | `internal/calc` |
| Application / service | `internal/service` | steps 1-6, operation lookup, orchestration | `net/http` |
| Core calculator | `internal/calc` | the registry, steps 7-10, pure arithmetic | `net/http`, `encoding/json` |

Dependencies point inward: `api` imports `service`, `service` imports `calc`, `calc` imports
nothing beyond `math` and `errors` (Constitution II).

The `api` layer owns the `code -> HTTP status` table, because status is a transport concern.
The core layer names failures; it does not know they become HTTP.
