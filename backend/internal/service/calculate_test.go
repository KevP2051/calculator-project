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
