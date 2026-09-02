package service

import (
	"encoding/json"
	"testing"
)

func numbers(literals ...string) []any {
	operands := make([]any, len(literals))
	for i, literal := range literals {
		operands[i] = json.Number(literal)
	}
	return operands
}

func TestCalculateReturnsResultsForBasicOperations(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		operands  []any
		want      float64
	}{
		{"addition", "add", numbers("2", "3"), 5},
		{"subtraction", "subtract", numbers("10", "4"), 6},
		{"subtraction respects operand order", "subtract", numbers("4", "10"), -6},
		{"multiplication", "multiply", numbers("6", "7"), 42},
		{"division", "divide", numbers("10", "4"), 2.5},
		{"division respects operand order", "divide", numbers("4", "10"), 0.4},
		{"decimal operands", "add", numbers("2.5", "0.25"), 2.75},
		{"negative operands", "multiply", numbers("-6", "7"), -42},
		{"representation noise is preserved", "add", numbers("0.1", "0.2"), 0.30000000000000004},
		{"exact cancellation", "subtract", numbers("5", "5"), 0},
		{"exponent notation", "multiply", numbers("1e2", "3"), 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calculate(Request{Operation: tt.operation, Operands: tt.operands})
			if err != nil {
				t.Fatalf("unexpected error: %+v", err)
			}
			if got != tt.want {
				t.Errorf("result = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateIsDeterministic(t *testing.T) {
	req := Request{Operation: "divide", Operands: numbers("10", "3")}

	first, err := Calculate(req)
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	second, err := Calculate(Request{Operation: "divide", Operands: numbers("10", "3")})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if first != second {
		t.Errorf("repeated request returned %v then %v", first, second)
	}
}

func TestCalculateRejectsOperandsOutsideTheRepresentableRange(t *testing.T) {
	tests := []struct {
		name    string
		operand string
	}{
		{"too large", "1e400"},
		{"too large and negative", "-1e400"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Calculate(Request{Operation: "add", Operands: numbers(tt.operand, "1")})
			if err == nil {
				t.Fatal("expected an error")
			}
			if err.Code != CodeOperandOutOfRange {
				t.Errorf("code = %q, want %q", err.Code, CodeOperandOutOfRange)
			}
		})
	}
}

func TestCalculateDoesNotRoundResults(t *testing.T) {
	got, err := Calculate(Request{Operation: "add", Operands: numbers("0.1", "0.2")})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if got == 0.3 {
		t.Fatal("0.1 + 0.2 returned 0.3, so the result was rounded")
	}
}

func expectFailure(t *testing.T, req Request) *Error {
	t.Helper()

	result, err := Calculate(req)
	if err == nil {
		t.Fatalf("expected a failure, got result %v", result)
	}
	if result != 0 {
		t.Errorf("result = %v, want 0 alongside a failure", result)
	}
	if err.Message == "" {
		t.Error("failure carries no message")
	}

	return err
}

func TestCalculateReportsMissingFields(t *testing.T) {
	tests := []struct {
		name string
		req  Request
	}{
		{"operation absent", Request{Operands: numbers("2", "3")}},
		{"operation empty", Request{Operation: "", Operands: numbers("2", "3")}},
		{"operands absent", Request{Operation: "add"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expectFailure(t, tt.req).Code; got != CodeMissingField {
				t.Errorf("code = %q, want %q", got, CodeMissingField)
			}
		})
	}
}

func TestCalculateRejectsUnsupportedOperations(t *testing.T) {
	for _, operation := range []string{"tangent", "ADD", "sum", "plus", " add"} {
		t.Run(operation, func(t *testing.T) {
			if got := expectFailure(t, Request{Operation: operation, Operands: numbers("2", "3")}).Code; got != CodeUnsupportedOperation {
				t.Errorf("code = %q, want %q", got, CodeUnsupportedOperation)
			}
		})
	}
}

func TestCalculateRejectsWrongOperandCount(t *testing.T) {
	tests := []struct {
		name string
		req  Request
	}{
		{"one operand for add", Request{Operation: "add", Operands: numbers("2")}},
		{"three operands for add", Request{Operation: "add", Operands: numbers("2", "3", "4")}},
		{"three operands for divide", Request{Operation: "divide", Operands: numbers("2", "3", "4")}},
		{"empty operand list", Request{Operation: "add", Operands: []any{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expectFailure(t, tt.req).Code; got != CodeInvalidOperandCount {
				t.Errorf("code = %q, want %q", got, CodeInvalidOperandCount)
			}
		})
	}
}

func TestCalculateRejectsOperandsThatAreNotNumbers(t *testing.T) {
	tests := []struct {
		name    string
		operand any
	}{
		{"text", "abc"},
		{"numeric text", "5"},
		{"the word NaN", "NaN"},
		{"the word Infinity", "Infinity"},
		{"the word -Infinity", "-Infinity"},
		{"boolean", true},
		{"null", nil},
		{"object", map[string]any{"value": json.Number("5")}},
		{"array", []any{json.Number("5")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := Request{Operation: "add", Operands: []any{tt.operand, json.Number("1")}}
			if got := expectFailure(t, req).Code; got != CodeInvalidOperand {
				t.Errorf("code = %q, want %q", got, CodeInvalidOperand)
			}
		})
	}
}

func TestNotANumberAndOutOfRangeAreDistinctCodes(t *testing.T) {
	notANumber := expectFailure(t, Request{Operation: "add", Operands: []any{"abc", json.Number("1")}})
	outOfRange := expectFailure(t, Request{Operation: "add", Operands: numbers("1e400", "1")})

	if notANumber.Code == outOfRange.Code {
		t.Fatalf("both failures used code %q", notANumber.Code)
	}
	if notANumber.Code != CodeInvalidOperand {
		t.Errorf("non-numeric operand code = %q, want %q", notANumber.Code, CodeInvalidOperand)
	}
	if outOfRange.Code != CodeOperandOutOfRange {
		t.Errorf("out-of-range operand code = %q, want %q", outOfRange.Code, CodeOperandOutOfRange)
	}
}

func TestCalculateReportsCalculationFailures(t *testing.T) {
	tests := []struct {
		name string
		req  Request
		want Code
	}{
		{"division by zero", Request{Operation: "divide", Operands: numbers("1", "0")}, CodeDivisionByZero},
		{"overflow", Request{Operation: "multiply", Operands: numbers("1e308", "10")}, CodeResultOverflow},
		{"underflow", Request{Operation: "multiply", Operands: numbers("1e-200", "1e-200")}, CodeResultUnderflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expectFailure(t, tt.req).Code; got != tt.want {
				t.Errorf("code = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidationHappensBeforeCalculation(t *testing.T) {
	req := Request{Operation: "divide", Operands: []any{"abc", json.Number("0")}}

	if got := expectFailure(t, req).Code; got != CodeInvalidOperand {
		t.Errorf("code = %q, want %q: the invalid operand must be caught before the zero divisor", got, CodeInvalidOperand)
	}
}
