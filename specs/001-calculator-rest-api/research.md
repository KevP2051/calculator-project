# Phase 0 Research: Calculator REST API

**Feature**: 001-calculator-rest-api | **Date**: 2026-09-01

All items below were unknowns in Technical Context. None remain unresolved.

---

## R1. Go version and HTTP routing

**Decision**: Go 1.22 or later, routing with `net/http`'s `ServeMux` using method+pattern
registration (`mux.HandleFunc("POST /api/v1/calculate", ...)`). No third-party router.

**Rationale**: Go 1.22 added method and wildcard patterns to the standard `ServeMux`, which
removes the last common reason to pull in `chi`, `gorilla/mux`, or `gin`. It also returns
`405 Method Not Allowed` automatically for a registered path called with the wrong method,
which the error contract needs. Satisfies Constitution IV (standard library preferred) with
zero justification burden.

**Alternatives considered**:

- `chi` — nice middleware ergonomics, but a dependency buying nothing at one endpoint.
- `gin` — brings a whole framework, JSON binding, and its own error conventions; directly
  contradicts Constitution IV and V.
- Pre-1.22 `ServeMux` with manual method checks — works, but requires hand-rolling the 405
  path that the standard library now provides.

---

## R2. Endpoint shape: one endpoint or one per operation

**Decision**: A single `POST /api/v1/calculate`, with the operation named in the request
body. Plus `GET /healthz`.

**Rationale**: Every operation shares the same validation pipeline, the same error envelope,
and the same status mapping. Seven endpoints would duplicate that seven times or force a
shared handler anyway. One endpoint gives the client one contract to learn, and adding an
operation later is a new entry in a map rather than a new route. It also makes FR-017
(identical requests produce identical responses) trivially observable.

**Alternatives considered**:

- `POST /api/v1/add`, `/subtract`, ... — more RESTful in appearance, but the resource being
  acted on is the calculation, not the operation; splitting duplicates validation and turns
  the unsupported-operation error (FR-008) into a router 404 rather than a domain error the
  contract defines.
- `GET /api/v1/calculate?op=add&a=1&b=2` — cacheable and curl-friendly, but pushes every
  operand through query-string parsing, where distinguishing "not a number" from "out of
  range" (FR-021 vs FR-022) gets murkier, and floats in URLs invite encoding trouble.

---

## R3. Request shape: named operands or an operand array

**Decision**: `{"operation": "add", "operands": [2, 3]}`. Square root takes a one-element
array.

**Rationale**: Arity is a first-class concept in the spec (FR-007a, FR-023). An array makes
the operand count directly observable, so the arity check is a single length comparison and
the invalid-operand-count error has an obvious source. It also gives one uniform request
shape for unary and binary operations rather than two.

**Alternatives considered**:

- `{"operation": "sqrt", "a": 9}` with an optional `b` — friendlier in a curl example, but
  "wrong operand count" becomes "count the recognized keys that are present", and it
  collides with the spec's rule that unrecognized extra fields are ignored: a `b` sent to
  `sqrt` is a *recognized* field that must still be an arity error, which is an awkward rule
  to state and to test.
- All operations binary with a sentinel second operand for square root — contradicts
  FR-007a.

---

## R4. Distinguishing malformed JSON, non-numbers, and out-of-range numbers

The sharpest implementation question in the feature. FR-019, FR-021, FR-022 and FR-022a
require three different error identifiers that Go's `encoding/json` collapses into a single
`*json.UnmarshalTypeError` if operands are decoded straight into `float64`.

**Decision**: Decode operands as `[]any` with `json.Decoder.UseNumber()`, then classify each
element in the service layer:

| Element after decode | Classification |
|---|---|
| `json.Number` that `strconv.ParseFloat` converts to a finite value | valid operand |
| `json.Number` where `ParseFloat` returns `strconv.ErrRange` (e.g. `1e400`) | `OPERAND_OUT_OF_RANGE` |
| anything else — `string`, `bool`, `nil`, object, array | `INVALID_OPERAND` |

Malformed JSON is caught earlier, by the decode call itself, and becomes `MALFORMED_JSON`.

**Rationale**: `UseNumber` keeps every JSON number as its original literal text instead of
eagerly converting to `float64`, so the out-of-range case survives as data rather than
becoming `±Inf` or a decode error. `strconv.ParseFloat` reports range overflow as a distinct
sentinel (`strconv.ErrRange`), which is exactly the FR-021/FR-022 split. Textual non-finite
forms need no special case: `"NaN"` and `"Infinity"` arrive as Go `string` values, not
`json.Number`, so they fall into `INVALID_OPERAND` by the same rule that catches `"abc"` —
satisfying FR-035 without a dedicated branch.

**Alternatives considered**:

- Decoding into `*float64` and string-matching the text of the returned
  `*json.UnmarshalTypeError` to tell "not a number" from "out of range" — works today, but
  couples the error contract to a message the Go team is free to reword.
- Decoding into `[]json.RawMessage` and inspecting the first byte of each token — precise,
  but hand-rolls token classification that `UseNumber` already does correctly.
- Checking `math.IsInf` after conversion — a useful belt-and-braces check, but insufficient
  alone: it cannot distinguish an operand that *arrived* as infinity from one that
  overflowed during parsing.

---

## R5. Detecting underflow to zero (FR-024a / FR-024b)

**Decision**: The rule is stated on the *result*: reject when the result is zero and every
operand was non-zero. One shared helper. It is called by multiplication, division,
exponentiation, and percentage — and by exactly those four, because they are the only
operations that can reach the condition.

**Rationale**: The operation list is derived, not chosen (FR-024d). Every finite binary64
value is an integer multiple of `2^-1074`, the smallest positive subnormal. The exact sum or
difference of two of them is therefore also a multiple of `2^-1074`, so a non-zero exact
result has magnitude at least `2^-1074` — exactly representable. Rounding cannot carry it to
zero, which would need a magnitude below `2^-1075`. **Addition and subtraction cannot
underflow to zero in binary64.** A zero from non-zero operands there is always exact
cancellation, so calling the helper for them would not add detection — it would
misclassify `5 - 5` as an error and violate FR-024b.

Square root is excluded for a separate reason: `sqrt(x) == 0` only when `x == 0`, since the
square root of the smallest positive subnormal is approximately `2.2e-162`.

For the remaining four, the exact result of non-zero operands is never zero, so a computed
zero can only be underflow. The helper is sound for exactly these four and unsound
elsewhere.

Negative zero needs no special handling: in Go, `-0.0 == 0` is true, so a result of `-0` is
caught by the same comparison.

**Verification**: checked empirically against binary64 arithmetic. A brute-force sweep of
2000x2000 subnormal operand pairs produced no addition or subtraction that returned zero
from operands whose exact difference was non-zero. Two apparent counterexamples turned out
to be decoys: `1e-320 - 9.99999e-321` returns `0` because both decimal literals round to the
*same* binary64 value (2024 x `2^-1074`), and `MIN_SUB - MIN_SUB*0.9` likewise reduces to
`MIN_SUB - MIN_SUB`. Both are exact cancellation. Meanwhile `1e-200 * 1e-200`,
`1e-300 / 1e300`, `2^-100000` and `(1e-200/100) * 1e-160` all return zero from non-zero
operands, confirming the four guarded operations do reach the condition.

**Alternatives considered**:

- Returning `0`, per IEEE 754's own gradual-underflow behavior — rejected by the developer
  during clarification.
- Applying the guard to all seven operations for symmetry — adds no detection capability and
  introduces a defect: `5 - 5` and `5 + -5` would return `RESULT_UNDERFLOW`, violating
  FR-024b.
- Documenting add/subtract as an unsupported underflow case — factually wrong, and it would
  invite a future maintainer to "fix" it by widening the guard into the bug above.
- Comparing against `math.SmallestNonzeroFloat64` to catch subnormal results too — goes
  beyond FR-024a, which triggers only at zero; subnormal results retain real magnitude and
  fall under the accepted-precision-loss rule.

---

## R6. Error identifier catalog and HTTP status mapping

**Decision**: Eleven identifiers across two client-meaningful status codes.

`400 Bad Request` — the request itself is wrong:
`MALFORMED_JSON`, `MISSING_FIELD`, `UNSUPPORTED_OPERATION`, `INVALID_OPERAND`,
`OPERAND_OUT_OF_RANGE`, `INVALID_OPERAND_COUNT`

`422 Unprocessable Entity` — the request is well-formed and understood, but the calculation
cannot be performed or represented:
`DIVISION_BY_ZERO`, `OPERAND_OUT_OF_DOMAIN`, `RESULT_OVERFLOW`, `RESULT_UNDERFLOW`,
`RESULT_UNDEFINED`

**Rationale**: The split is the one distinction a client actually acts on: 400 means "fix
what you sent", 422 means "what you sent was fine, but that calculation has no answer".
FR-012a requires one identifier per category and no identifier serving two categories; this
catalog satisfies that. `MISSING_FIELD` covers both an absent `operation` and an absent
`operands` because FR-020 defines them as a single category. `404` and `405` come from the
router for an unknown path or wrong method and sit outside this catalog, as they are not
calculation failures.

**Alternatives considered**:

- `422` for everything non-malformed — loses the "your input is broken" versus "your math is
  undefined" distinction that SC-005 exists to guarantee.
- `400` for everything — same loss, and treats a mathematically undefined but perfectly
  well-formed request as a client syntax error.
- `500` for overflow — would violate the spec's rule that no defined failure reaches an
  unhandled server error.

---

## R7. Exponentiation domain errors

**Decision**: Reject a negative base raised to a non-integer exponent *before* calling
`math.Pow`, as `OPERAND_OUT_OF_DOMAIN`. Let `0` raised to a negative exponent go through
`math.Pow` and be caught as `RESULT_OVERFLOW` by the non-finite result check.

**Rationale**: `math.Pow(-8, 1.0/3.0)` returns `NaN`, which would otherwise be reported as
`RESULT_UNDEFINED`. The spec's Edge Cases section classifies this as a *domain* error,
because the true result is a complex number the service does not model — a fact about the
inputs, not about the computation. Checking `base < 0 && exp != math.Trunc(exp)` up front
produces the right identifier. Conversely `math.Pow(0, -1)` returns `+Inf`, which the spec
does classify as unrepresentable, so no pre-check is needed there.

**Alternatives considered**:

- Mapping every `NaN` result to `RESULT_UNDEFINED` — simpler, but contradicts the spec's
  explicit edge-case classification and would leave `OPERAND_OUT_OF_DOMAIN` used only by
  square root.

---

## R8. Test comparison strategy

**Decision**: Table-driven tests in the standard `testing` package. Comparison follows
FR-031a: exact `==` where the expected value is exactly representable in binary64, relative
`1e-9` only for irrational and transcendental results. One dedicated test asserts
`0.1 + 0.2 == 0.30000000000000004` exactly, per FR-031b.

**Rationale**: The exact tier is the only thing that can detect a violation of FR-017a's
no-rounding guarantee — a tolerance comparison would pass a service that silently rounded to
`0.3`. Table-driven tests are idiomatic Go and keep one row per spec edge case, which makes
the mapping from spec to suite auditable during review.

**Alternatives considered**:

- `testify/assert` — a dependency for what `if got != want { t.Errorf(...) }` already does;
  fails Constitution IV.
- `math.Abs(got-want) < 1e-9` everywhere — hides rounding bugs, as above.

---

## R9. CORS and configuration

**Decision**: One small middleware allowing a single configurable origin, read from the
`CORS_ORIGIN` environment variable, defaulting to `http://localhost:5173` (the Vite dev
server). Listen address from `PORT`, defaulting to `8080`. Nothing else is configurable.

**Rationale**: The frontend is served from a different origin during development, so without
CORS headers the browser blocks every call — a functional requirement of the deliverable,
not speculative flexibility. Two environment variables with sane defaults is the smallest
thing that works; Constitution V forbids a configuration *system*, not two `os.Getenv` calls
with defaults.

**Alternatives considered**:

- `rs/cors` — a dependency for roughly fifteen lines of header setting.
- Hardcoding `*` — acceptable for a public calculator with no credentials, but a poor habit
  to demonstrate in an assessment, and it needs changing the moment anything is added.
- Hardcoding the dev origin with no env var — breaks as soon as the frontend is served from
  a container or a different port.

---

## R10. Containerization

**Decision**: Out of scope for this feature. Not planned, not tasked.

**Rationale**: Constitution V lists containerization among the things that MUST NOT be
introduced unless the developer explicitly requests them. The assignment marks a Dockerfile
optional, and no explicit request has been made. Recorded here so the decision is visible
rather than forgotten.

**If the developer requests it later**: a two-stage `golang:1.22` to
`gcr.io/distroless/static` build for the backend, an `nginx:alpine` stage serving the built
frontend, and a `docker-compose.yml` wiring the two with `CORS_ORIGIN` pointed at the
frontend container. Additive, requiring no rework of anything planned here.
