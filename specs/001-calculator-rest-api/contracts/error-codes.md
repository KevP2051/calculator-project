# Error Contract: Calculator REST API

**Feature**: 001-calculator-rest-api | **Date**: 2026-09-01

Every failure returns the same envelope. `code` is stable and is the only field a client
should branch on (FR-012). `message` is for developers and may be reworded without notice.

```json
{
  "error": {
    "code": "DIVISION_BY_ZERO",
    "message": "cannot divide by zero"
  }
}
```

No error response ever carries a numeric result (FR-015).

---

## Catalog

Eleven identifiers. One per failure category defined in the spec, none shared (FR-012a).

### 400 Bad Request — the request itself is wrong

| Code | Fires when | Spec |
|---|---|---|
| `MALFORMED_JSON` | Body is not well-formed JSON, or is not a JSON object | FR-019 |
| `MISSING_FIELD` | `operation` absent or empty, or `operands` absent | FR-020 |
| `UNSUPPORTED_OPERATION` | `operation` names something the service does not implement | FR-008 |
| `INVALID_OPERAND_COUNT` | `operands` length does not match the operation's arity | FR-023, FR-007a |
| `INVALID_OPERAND` | An operand is not a JSON number — text, boolean, null, object, array, or a textual non-finite form such as `"NaN"` or `"Infinity"` | FR-021, FR-035 |
| `OPERAND_OUT_OF_RANGE` | An operand is a well-formed number whose magnitude exceeds the finite binary64 range, such as `1e400` | FR-022 |

`INVALID_OPERAND` and `OPERAND_OUT_OF_RANGE` split on cause, not on type: the first means
"that was not a number", the second means "that was a number I cannot represent" (FR-022a).

### 422 Unprocessable Entity — well-formed, but the calculation has no answer

| Code | Fires when | Spec |
|---|---|---|
| `DIVISION_BY_ZERO` | Divisor is zero | FR-025 |
| `OPERAND_OUT_OF_DOMAIN` | Square root of a negative operand, or a negative base raised to a non-integer exponent | FR-026, Edge Cases |
| `RESULT_OVERFLOW` | Result is infinite, including `0` raised to a negative exponent | FR-024 |
| `RESULT_UNDERFLOW` | Result is zero from all-non-zero operands. Reachable only in multiply, divide, power and percentage — addition and subtraction cannot underflow to zero in binary64 (FR-024d) | FR-024a |
| `RESULT_UNDEFINED` | Result is NaN | FR-024 |

`RESULT_OVERFLOW`, `RESULT_UNDERFLOW` and `RESULT_UNDEFINED` are three distinct categories
with three distinct identifiers, so a client can tell "too large" from "too small" from
"undefined" without reading the message (FR-024c).

### Router-level responses

Not part of the catalog, because they are not calculation failures:

| Status | When |
|---|---|
| `404 Not Found` | Unknown path |
| `405 Method Not Allowed` | Known path, wrong method — emitted by `net/http`'s `ServeMux` |

---

## Cases that are NOT errors

| Case | Behavior | Spec |
|---|---|---|
| Extra unrecognized fields in the request body | Ignored | Edge Cases |
| `0.1 + 0.2` | Returns `0.30000000000000004` exactly, unrounded | FR-017a, FR-031b |
| Precision loss within the representable range | Result returned as computed | Assumptions |
| `5 - 5`, `5 + -5` | Returns `0` — exact cancellation, not underflow | FR-024b |
| Subtraction of adjacent or equal subnormals | Returns a subnormal or `0`, always a success — binary64 addition and subtraction cannot underflow to zero | FR-024d, FR-031c |
| `0 * 5`, `0 / 5`, `percentage(0, 5)`, `percentage(5, 0)`, `0 ** 5` | Returns `0` — a zero operand, not underflow | FR-024b |
| `sqrt(0)` | Returns `0` — square root cannot underflow | FR-024b |
| `0 ** 0` | Returns `1`, per floating-point convention | Edge Cases |
| Negative operands for any operation except square root | Valid input | Edge Cases |
