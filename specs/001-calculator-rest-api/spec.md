# Feature Specification: Calculator REST API

**Feature Directory**: `specs/001-calculator-rest-api`

**Created**: 2026-09-01

**Status**: Draft

**Input**: User description: "Calculator backend exposing addition, subtraction, multiplication, division, exponentiation, square root and percentage through a stateless REST API returning JSON results and consistent machine-readable JSON errors, with defined behavior for arithmetic and input edge cases."

## Clarifications

### Session 2026-09-01

- Q: What does the percentage operation compute, and how many operands does it take? → A: Binary — `percentage(a, b) = (a / 100) * b`, e.g. `percentage(15, 200) = 30`. All binary operations take exactly two operands; square root is the only unary operation.
- Q: Should results with floating-point representation noise be returned exactly or rounded? → A: Exactly, unrounded. The backend does not format results; the client owns display rounding. Acceptance tests compare within a small tolerance. (Tolerance later fixed to the two-tier rule in FR-031a.)
- Q: Should the spec name IEEE 754 binary64 (float64) as the numeric model, or leave the numeric type to the plan? → A: Name it normatively in the spec. IEEE 754 binary64 is a language-neutral published standard and a client-observable contract property (which operand values are accepted, what precision is returned), not a technology choice. See the Numeric Model requirements.
- Q: Is a syntactically valid but unrepresentable number such as `1e400` an out-of-range-input error or an invalid-operand error? → A: Out-of-range-input. The two categories keep distinct error identifiers and split on cause: a value that parses as a number but exceeds the finite binary64 range is out-of-range-input; a value that is not a number at all is invalid-operand.
- Q: Should a calculation that underflows to zero (for example `1e-200 * 1e-200`) return 0 as a success or be rejected? → A: Rejected as an underflow-result error, a category distinct from overflow. Applies only to multiplication, division, exponentiation and percentage, and only when every operand is non-zero — a zero produced by addition, subtraction, or a zero operand is a legitimate result.
- Q: Should underflow carry its own error identifier or share the overflow identifier? → A: Its own. Overflow, underflow, and undefined (NaN) results are three distinct categories with three distinct identifiers, as FR-012 already requires; the client remedy differs for each.
- Q: What tolerance should acceptance tests use when comparing a computed result to the expected value? → A: Two-tier. Exact equality where the expected value is exactly representable in binary64 (`2 + 3`, `0.1 + 0.2 == 0.30000000000000004`, `percentage(15, 200) == 30`); relative tolerance of `1e-9` only where the true value is irrational or transcendental (square root, non-integer exponentiation). The exact tier is what makes FR-017a's no-rounding guarantee testable.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Perform a basic arithmetic calculation (Priority: P1)

A client application sends a calculation request naming one of the four basic operations
(addition, subtraction, multiplication, division) together with its operands, and receives
the computed result in JSON.

**Why this priority**: This is the minimum viable calculator. Without it the service has no
reason to exist; with only this, a client can already deliver a working four-function
calculator to end users.

**Independent Test**: Send requests for each of the four basic operations with known
operands and confirm the returned result matches the expected value. Delivers value as a
usable four-function calculator backend.

**Acceptance Scenarios**:

1. **Given** the service is running, **When** a client requests addition of two valid
   numbers, **Then** the response is a success response containing their sum.
2. **Given** the service is running, **When** a client requests subtraction of two valid
   numbers, **Then** the response is a success response containing their difference, in the
   order supplied.
3. **Given** the service is running, **When** a client requests multiplication of two valid
   numbers, **Then** the response is a success response containing their product.
4. **Given** the service is running, **When** a client requests division of two valid
   numbers and the divisor is not zero, **Then** the response is a success response
   containing their quotient, in the order supplied.
5. **Given** the service is running, **When** a client sends the same request twice,
   **Then** both responses are identical and neither request affects the other.

---

### User Story 2 - Receive predictable, machine-readable errors (Priority: P2)

A client application sends a request that cannot be fulfilled — malformed input, a missing
field, an unknown operation, a non-numeric operand, or an arithmetically undefined
operation — and receives a JSON error response it can branch on programmatically without
parsing human-readable prose.

**Why this priority**: A calculator that succeeds but fails opaquely cannot be integrated. A
client needs to distinguish "you sent bad input" from "that calculation is undefined" in
order to show the right message to an end user.

**Independent Test**: Send one request per failure category and confirm each response uses
the same error envelope shape, carries a distinct stable error identifier, and never
returns a numeric result.

**Acceptance Scenarios**:

1. **Given** the service is running, **When** a client requests division with a divisor of
   zero, **Then** the response is an error identifying the operation as undefined, and no
   numeric result is returned.
2. **Given** the service is running, **When** a client sends a request body that is not
   valid JSON, **Then** the response is an error in the standard error format identifying
   the input as malformed.
3. **Given** the service is running, **When** a client omits a required field, **Then** the
   response is an error identifying what is missing.
4. **Given** the service is running, **When** a client names an operation the service does
   not support, **Then** the response is an error identifying the operation as unsupported.
5. **Given** the service is running, **When** any error occurs, **Then** the error response
   has the same overall structure as every other error response, including a stable error
   identifier distinct from other error categories and a human-readable message.
6. **Given** the service is running, **When** input is invalid, **Then** no calculation is
   attempted before the input is rejected.

---

### User Story 3 - Perform extended operations (Priority: P3)

A client application requests exponentiation, square root, or percentage and receives the
computed result, with undefined or unrepresentable cases reported as errors rather than as
special numeric values.

**Why this priority**: Additive to the core calculator. The service is already valuable
without these, but the assignment requires them and they exercise the unary-operand and
domain-error paths.

**Independent Test**: Send requests for each extended operation with known operands and
confirm results, then send the domain-invalid cases (square root of a negative number,
exponentiation producing a non-finite result) and confirm they are reported as errors.

**Acceptance Scenarios**:

1. **Given** the service is running, **When** a client requests exponentiation with a valid
   base and exponent whose result is finite and representable, **Then** the response
   contains that result.
2. **Given** the service is running, **When** a client requests the square root of a
   non-negative number, **Then** the response contains its non-negative root.
3. **Given** the service is running, **When** a client requests the square root of a
   negative number, **Then** the response is an error identifying the input as outside the
   domain of the operation, and no numeric result is returned.
4. **Given** the service is running, **When** a client requests the percentage operation
   with operands 15 and 200, **Then** the response contains the result 30.
5. **Given** the service is running, **When** a client supplies the wrong number of operands
   for the requested operation, **Then** the response is an error identifying the operand
   count as invalid.

---

### Edge Cases

- **Division by zero**: Rejected as an undefined-operation error. No result, no infinity.
- **Square root of a negative number**: Rejected as a domain error. No result, no NaN.
- **Exponentiation producing a non-finite result** (very large base and exponent, or zero
  raised to a negative exponent): Rejected as an unrepresentable-result error.
- **Exponentiation of a negative base by a non-integer exponent**: The true result is not a
  real number; rejected as a domain error.
- **Zero raised to the power of zero**: Returns 1, matching standard floating-point
  convention.
- **Malformed request body**: Rejected as a malformed-input error before individual fields
  are inspected.
- **Missing operation or missing operand**: Rejected as a missing-input error naming what is
  absent.
- **Unsupported or unrecognized operation name**: Rejected as an unsupported-operation
  error. The service does not silently fall back to a default operation.
- **Operand that is not a number** (text, boolean, null, empty, or a textual non-finite
  form such as `"NaN"` or `"Infinity"`): Rejected as an invalid-operand error.
- **Operand outside the finite binary64 range** (a well-formed number such as `1e400` or
  `-1e400`): Rejected as an out-of-range-input error, before any calculation. This is a
  distinct error category from a value that is not a number at all.
- **Operand supplied as a non-finite value** (infinity, NaN, or their textual forms):
  Rejected as an invalid-operand error, since no such value is a number in JSON.
- **A calculation on valid operands whose result overflows the representable range**:
  Rejected as an unrepresentable-result error rather than returned as infinity.
- **A calculation whose result is undefined (NaN)**: Rejected as an undefined-result error.
- **Loss of precision within the representable range** (very small differences between very
  large numbers): Accepted and returned. Precision loss inherent to binary64 is not an
  error. The one exception is underflow to zero, below.
- **A calculation that underflows to zero** (for example `1e-200 * 1e-200`, or
  `1e-300 / 1e300`): Rejected as an underflow-result error, a category distinct from
  overflow (FR-024c). The condition is that the
  operation is multiplication, division, exponentiation, or percentage, every operand is
  non-zero, and the computed result is zero.
- **A zero result that is mathematically correct**: Accepted and returned. This covers exact
  cancellation in addition and subtraction (`5 - 5`, `5 + -5`), any operation with a zero
  operand (`0 * 5`, `0 / 5`, `percentage(0, 5)`, `percentage(5, 0)`, `0 ** 5`), and the
  square root of zero. These are never treated as underflow.
- **A result carrying floating-point representation noise** (for example, adding 0.1 and 0.2
  computing to 0.30000000000000004): Returned exactly as computed, unrounded. Not an error,
  and not corrected to the mathematically tidy value.
- **Percentage with a second operand of zero**: Valid. `percentage(a, 0)` returns 0. The
  percentage operation divides only by the constant 100, so it has no divide-by-zero case.
- **Percentage producing an unrepresentable result** (both operands very large): Rejected as
  an unrepresentable-result error, like any other overflow.
- **Wrong operand count for the operation** (two operands for square root, one for any other
  operation): Rejected as an invalid-operand-count error.
- **Extra or unrecognized fields in the request**: Ignored. Extra fields are not an error.
- **Negative operands** for any operation other than square root: Accepted as valid input.
- **Zero as an operand** anywhere other than as a divisor: Accepted as valid input.

## Requirements *(mandatory)*

### Functional Requirements

#### Operations

- **FR-001**: System MUST support addition of two operands and return their sum.
- **FR-002**: System MUST support subtraction of two operands and return their difference,
  in the order supplied.
- **FR-003**: System MUST support multiplication of two operands and return their product.
- **FR-004**: System MUST support division of two operands and return their quotient, in the
  order supplied.
- **FR-005**: System MUST support exponentiation of a base operand by an exponent operand
  and return the resulting power.
- **FR-006**: System MUST support the square root of a single operand and return its
  non-negative root.
- **FR-007**: System MUST support a percentage operation taking two operands, computing the
  first operand as a percentage of the second: `percentage(a, b) = (a / 100) * b`. For
  example, `percentage(15, 200)` returns `30`.
- **FR-007a**: Every operation except square root MUST take exactly two operands. Square root
  MUST take exactly one operand.
- **FR-008**: System MUST reject any operation it does not support, and MUST NOT substitute a
  default operation.

#### Numeric Model

- **FR-032**: The system's numeric model MUST be IEEE 754 binary64 (double-precision
  floating point, commonly `float64`). Every operand and every result is a value in that
  representation, and all range, precision, and finiteness rules below derive from it.
- **FR-033**: System MUST accept non-integer decimal operand values, not only whole numbers.
- **FR-034**: System MUST accept any finite binary64 value as an operand — the full
  representable range, positive, negative, and zero — subject only to the per-operation
  domain restrictions stated elsewhere in this specification (a divisor MUST NOT be zero per
  FR-025; a square-root operand MUST NOT be negative per FR-026).
- **FR-035**: System MUST reject NaN and positive or negative infinity as operand values,
  in numeric or textual form (invalid-operand, per FR-021).
- **FR-036**: System MUST NOT impose a maximum number of digits, decimal places, or operand
  magnitude beyond the limits binary64 itself imposes. This is the numeric case of the
  general rule in FR-027.

Non-finite *results* are governed by FR-016 and FR-024; unrounded return of results,
including binary64 representation noise, is governed by FR-017a.

#### API Behavior

- **FR-009**: System MUST expose calculator functionality over a REST API that accepts
  calculation requests from clients.
- **FR-010**: System MUST return successful results as JSON.
- **FR-011**: System MUST return errors as JSON in a single consistent structure used by
  every error response, regardless of failure category.
- **FR-012**: Every error response MUST carry a stable machine-readable error identifier that
  distinguishes its failure category from every other failure category defined in this
  specification, so a client can branch on it without parsing prose.
- **FR-012a**: Each failure category defined in this specification MUST map to exactly one
  error identifier, and no identifier may serve two categories. The literal identifier
  strings and their HTTP status mapping are part of the API contract and are fixed during
  planning, not here; this specification fixes only which categories exist and that they are
  distinguishable.
- **FR-013**: Every error response MUST carry a human-readable message describing the
  failure.
- **FR-014**: System MUST be stateless: a request outcome MUST depend only on that request,
  MUST NOT be affected by any previous request, and MUST NOT affect any subsequent one.
- **FR-015**: System MUST NOT return a numeric result in any error response.
- **FR-016**: System MUST NOT return a non-finite value (infinity or NaN) as a successful
  result under any circumstances.
- **FR-017**: Identical requests MUST produce identical responses.
- **FR-017a**: System MUST return the exact computed value of a successful calculation. It
  MUST NOT round, truncate, or otherwise reformat the result for presentation, and MUST NOT
  accept a client-supplied rounding precision. Display formatting is the client's
  responsibility.

#### Validation

- **FR-018**: System MUST validate the entire request — structure, operation name, operand
  presence, operand count, and operand values — before performing any calculation.
- **FR-019**: System MUST reject a request whose body is not well-formed JSON.
- **FR-020**: System MUST reject a request missing the operation or any operand the operation
  requires.
- **FR-021**: System MUST reject, as an invalid-operand error, an operand that is not a
  number at all — text, boolean, null, an empty value, or a textual non-finite form such as
  `"NaN"` or `"Infinity"`.
- **FR-022**: System MUST reject, as an out-of-range-input error, an operand that is a
  well-formed number but whose magnitude exceeds the finite binary64 range, such as `1e400`.
- **FR-022a**: The invalid-operand and out-of-range-input categories MUST use distinct
  stable error identifiers and MUST be distinguished by cause: FR-021 covers values that are
  not numbers; FR-022 covers numbers that are not representable.
- **FR-023**: System MUST reject a request whose operand count does not match the arity of
  the requested operation.
- **FR-024**: System MUST reject a calculation whose result is not finite, distinguishing an
  out-of-range (overflow) result from an undefined (NaN) result.
- **FR-024a**: System MUST reject, as an underflow-result error, a calculation that
  underflows to zero: a multiplication, division, exponentiation, or percentage whose
  operands are all non-zero but whose result is zero. The true result of such a calculation
  is a non-zero value too small to represent in binary64, and returning zero would report a
  magnitude the service did not compute.
- **FR-024b**: A zero result MUST be returned as a success whenever it is mathematically
  correct rather than the product of underflow — specifically, any zero produced by addition
  or subtraction (including exact cancellation such as `5 - 5`), and any zero produced by an
  operation for which at least one operand is zero (such as `0 * 5`, `0 / 5`, or
  `percentage(0, 5)`). Square root cannot underflow: its result is zero only when its
  operand is zero.
- **FR-024c**: Overflow (a non-finite result), underflow (FR-024a), and an undefined result
  (NaN) are three distinct failure categories and MUST carry three distinct stable error
  identifiers, per FR-012. A client MUST be able to tell "the result was too large" from
  "the result was too small" from "the result is undefined" using the identifier alone.
- **FR-025**: System MUST reject division when the divisor is zero.
- **FR-026**: System MUST reject the square root of a negative operand.
- **FR-027**: System MUST NOT impose validation rules beyond those required by this
  specification for correctness or safety — in particular, no arbitrary caps on operand
  magnitude, precision, or value beyond what binary64 itself imposes.

#### Testability

- **FR-028**: Core calculation behavior MUST be exercisable in automated tests without
  starting an HTTP server or issuing HTTP requests.
- **FR-029**: Each supported operation MUST have automated tests covering correct results for
  valid inputs.
- **FR-030**: Each error category defined in this specification MUST have at least one
  automated test asserting that the correct error identifier is returned.
- **FR-031**: The API success path and each distinct API error path MUST be exercisable in
  automated tests.
- **FR-031a**: Result comparison in automated tests MUST use exact equality wherever the
  mathematically expected value is exactly representable in binary64, and MUST use a
  relative tolerance of `1e-9` only where the true value is irrational or transcendental
  (square root of a non-perfect square, non-integer exponentiation).
- **FR-031b**: At least one test MUST assert exact equality on a result carrying
  representation noise — for example that adding `0.1` and `0.2` returns
  `0.30000000000000004` and not `0.3` — since a tolerance-based comparison cannot detect a
  violation of FR-017a.

### Key Entities

- **Calculation Request**: What the client asks for. Identifies one operation and carries the
  operands that operation requires. Carries no client identity, session, or history.
- **Calculation Result**: The outcome of a successful calculation. A single finite binary64
  value.
- **Calculation Error**: The outcome of a failed request. Carries a stable category
  identifier and a human-readable message. Never carries a numeric result.
- **Operation**: A named calculation with a fixed arity (how many operands it takes) and a
  domain (which operand values are valid for it).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All seven operations (addition, subtraction, multiplication, division,
  exponentiation, square root, percentage) return correct results for valid inputs, with
  100% of the operation acceptance scenarios in this specification passing. Correctness is
  judged against the mathematically expected value by exact equality where that value is
  exactly representable in binary64, and within a relative tolerance of `1e-9` where the true
  value is irrational or transcendental (FR-031a).
- **SC-002**: 100% of the edge cases listed in this specification produce the documented
  behavior, verified by automated tests.
- **SC-003**: Zero requests return a non-finite value (infinity or NaN) as a successful
  result.
- **SC-004**: Zero invalid requests result in a calculation being attempted; every invalid
  request is rejected during validation.
- **SC-005**: A client can distinguish every error category defined in this specification
  using only the machine-readable error identifier, without inspecting the human-readable
  message.
- **SC-006**: 100% of error responses share the same structure.
- **SC-007**: Core calculation behavior is covered by automated tests that run without
  starting the HTTP layer.
- **SC-007a**: At least one automated test proves results are returned unrounded, by
  asserting exact equality on a result that carries binary64 representation noise.
- **SC-008**: Repeating any request from the acceptance scenarios produces an identical
  response, confirming statelessness.
- **SC-009**: A client developer can integrate against the API using only the documented
  contract, without reading the service source.

## Assumptions

- **Arity**: Addition, subtraction, multiplication, division, exponentiation, and percentage
  take exactly two operands. Square root is the only unary operation, taking exactly one.
- **Single calculation per request**: Each request carries exactly one calculation. Batch and
  chained (more than two operands) calculations are out of scope; a client composes them by
  issuing successive requests.
- **Operand order matters** for subtraction, division, and exponentiation; operands are
  interpreted in the order the client supplies them.
- **Real numbers only**: The service works in real numbers. Operations whose true result is
  complex (square root of a negative number, negative base raised to a non-integer exponent)
  are domain errors, not results.
- **Precision is best-effort, and results are unrounded**: Results are subject to the
  inherent precision limits of IEEE 754 binary64 (roughly 15-17 significant decimal digits),
  with the single exception that a total loss of magnitude — underflow to zero — is an error
  rather than an accepted approximation (FR-024a). Representation noise within those limits
  is expected behavior, not a defect, and is returned as computed rather than corrected.
  Acceptance tests therefore compare results exactly where the expected value is exactly
  representable, and within a relative tolerance of `1e-9` only for irrational and
  transcendental results. Arbitrary-precision and exact decimal arithmetic are out of scope,
  as is any result formatting or rounding.
- **No authentication, authorization, rate limiting, quotas, or per-client accounting** —
  none were requested and none are needed for correctness or safety at this scope.
- **No persistence, calculation history, or audit log** — the service is stateless by
  requirement.
- **Frontend and UI concerns are out of scope**, per the project constitution. This
  specification defines only behavior a client can observe through the API.
- **Consumers are client applications**, not end users directly. "Client" in this document
  means the application integrating with the API.
- **Deployment, hosting, and operational concerns** (scaling, monitoring, availability
  targets) are out of scope for this feature.
