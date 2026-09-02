package calc

import (
	"errors"
	"math"
	"testing"
)

func compute(t *testing.T, name string, operands ...float64) (float64, error) {
	t.Helper()

	op, ok := Lookup(name)
	if !ok {
		t.Fatalf("operation %q is not registered", name)
	}

	return Compute(op, operands)
}

func TestBasicOperationsAreRegisteredWithTheirArityAndUnderflowSetting(t *testing.T) {
	tests := []struct {
		name           string
		arity          int
		checkUnderflow bool
	}{
		{"add", 2, false},
		{"subtract", 2, false},
		{"multiply", 2, true},
		{"divide", 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, ok := Lookup(tt.name)
			if !ok {
				t.Fatalf("operation %q is not registered", tt.name)
			}
			if op.Arity != tt.arity {
				t.Errorf("Arity = %d, want %d", op.Arity, tt.arity)
			}
			if op.CheckUnderflow != tt.checkUnderflow {
				t.Errorf("CheckUnderflow = %v, want %v", op.CheckUnderflow, tt.checkUnderflow)
			}
		})
	}
}

func TestBasicOperationResults(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		operands  []float64
		want      float64
	}{
		{"addition", "add", []float64{2, 3}, 5},
		{"addition of negatives", "add", []float64{-2, -3}, -5},
		{"addition of decimals", "add", []float64{2.5, 0.25}, 2.75},
		{"subtraction", "subtract", []float64{10, 4}, 6},
		{"subtraction respects operand order", "subtract", []float64{4, 10}, -6},
		{"multiplication", "multiply", []float64{6, 7}, 42},
		{"multiplication by a negative", "multiply", []float64{6, -7}, -42},
		{"division", "divide", []float64{10, 4}, 2.5},
		{"division respects operand order", "divide", []float64{4, 10}, 0.4},
		{"representation noise is preserved", "add", []float64{0.1, 0.2}, 0.30000000000000004},
		{"exact cancellation by subtraction", "subtract", []float64{5, 5}, 0},
		{"exact cancellation by addition", "add", []float64{5, -5}, 0},
		{"zero first operand in multiplication", "multiply", []float64{0, 5}, 0},
		{"zero second operand in multiplication", "multiply", []float64{5, 0}, 0},
		{"zero numerator in division", "divide", []float64{0, 5}, 0},
		{"equal subnormals cancel to zero", "subtract", []float64{5e-324, 5e-324}, 0},
		{"adjacent subnormals leave a subnormal", "subtract", []float64{1.5e-323, 1e-323}, math.SmallestNonzeroFloat64},
		{"largest finite value is not overflow", "add", []float64{math.MaxFloat64, 0}, math.MaxFloat64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compute(t, tt.operation, tt.operands...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("%s(%v) = %v, want %v", tt.operation, tt.operands, got, tt.want)
			}
		})
	}
}

func TestBasicOperationFailures(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		operands  []float64
		want      error
	}{
		{"division by zero", "divide", []float64{1, 0}, ErrDivisionByZero},
		{"division of zero by zero", "divide", []float64{0, 0}, ErrDivisionByZero},
		{"multiplication overflows", "multiply", []float64{1e308, 10}, ErrOverflow},
		{"addition overflows", "add", []float64{math.MaxFloat64, math.MaxFloat64}, ErrOverflow},
		{"multiplication underflows", "multiply", []float64{1e-200, 1e-200}, ErrUnderflow},
		{"division underflows", "divide", []float64{1e-300, 1e300}, ErrUnderflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compute(t, tt.operation, tt.operands...)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if got != 0 {
				t.Errorf("result = %v, want 0 alongside an error", got)
			}
		})
	}
}

func TestNoiseIsNotRounded(t *testing.T) {
	got, err := compute(t, "add", 0.1, 0.2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == 0.3 {
		t.Fatal("0.1 + 0.2 returned 0.3, so the result was rounded")
	}
	if got != 0.30000000000000004 {
		t.Errorf("0.1 + 0.2 = %v, want 0.30000000000000004", got)
	}
}

func TestSubtractionCannotUnderflowToZero(t *testing.T) {
	tests := []struct {
		name     string
		operands []float64
	}{
		{"equal subnormals", []float64{5e-324, 5e-324}},
		{"adjacent subnormals", []float64{1.5e-323, 1e-323}},
		{"tiny difference between large values", []float64{1e308, 1e308}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := compute(t, "subtract", tt.operands...); err != nil {
				t.Errorf("subtract(%v) returned %v, want a successful result", tt.operands, err)
			}
		})
	}
}

func TestPowerIsRegisteredWithItsArityAndUnderflowSetting(t *testing.T) {
	op, ok := Lookup("power")
	if !ok {
		t.Fatal("power is not registered")
	}
	if op.Arity != 2 {
		t.Errorf("Arity = %d, want 2", op.Arity)
	}
	if !op.CheckUnderflow {
		t.Error("CheckUnderflow = false, want true")
	}
}

func TestPowerResults(t *testing.T) {
	tests := []struct {
		name     string
		operands []float64
		want     float64
	}{
		{"positive integer exponent", []float64{2, 10}, 1024},
		{"exponent of zero", []float64{2, 0}, 1},
		{"zero to the power of zero", []float64{0, 0}, 1},
		{"zero base", []float64{0, 5}, 0},
		{"negative base with an odd integer exponent", []float64{-2, 3}, -8},
		{"negative base with an even integer exponent", []float64{-2, 4}, 16},
		{"negative integer exponent", []float64{2, -3}, 0.125},
		{"base of one", []float64{1, 1000}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compute(t, "power", tt.operands...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("power(%v) = %v, want %v", tt.operands, got, tt.want)
			}
		})
	}
}

func TestPowerFailures(t *testing.T) {
	tests := []struct {
		name     string
		operands []float64
		want     error
	}{
		{"zero to a negative exponent", []float64{0, -1}, ErrOverflow},
		{"result overflows", []float64{10, 400}, ErrOverflow},
		{"result underflows", []float64{2, -100000}, ErrUnderflow},
		{"negative base with a fractional exponent", []float64{-8, 1.0 / 3.0}, ErrOutOfDomain},
		{"negative base with a half exponent", []float64{-4, 0.5}, ErrOutOfDomain},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compute(t, "power", tt.operands...)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if got != 0 {
				t.Errorf("result = %v, want 0 alongside an error", got)
			}
		})
	}
}

func TestPowerReportsComplexResultsAsDomainErrorsNotUndefined(t *testing.T) {
	_, err := compute(t, "power", -8, 1.0/3.0)

	if errors.Is(err, ErrUndefined) {
		t.Fatal("a negative base with a fractional exponent reported an undefined result, want a domain error")
	}
	if !errors.Is(err, ErrOutOfDomain) {
		t.Errorf("error = %v, want %v", err, ErrOutOfDomain)
	}
}
