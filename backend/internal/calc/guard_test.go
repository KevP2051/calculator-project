package calc

import (
	"errors"
	"math"
	"testing"
)

var negativeZero = math.Copysign(0, -1)

func TestCheckFinite(t *testing.T) {
	tests := []struct {
		name   string
		result float64
		want   error
	}{
		{"ordinary value", 42.5, nil},
		{"zero", 0, nil},
		{"negative zero", negativeZero, nil},
		{"largest finite", math.MaxFloat64, nil},
		{"smallest subnormal", math.SmallestNonzeroFloat64, nil},
		{"positive infinity", math.Inf(1), ErrOverflow},
		{"negative infinity", math.Inf(-1), ErrOverflow},
		{"not a number", math.NaN(), ErrUndefined},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkFinite(tt.result); !errors.Is(got, tt.want) {
				t.Errorf("checkFinite(%v) = %v, want %v", tt.result, got, tt.want)
			}
		})
	}
}

func TestCheckUnderflow(t *testing.T) {
	tests := []struct {
		name     string
		result   float64
		operands []float64
		want     error
	}{
		{"zero from non-zero operands", 0, []float64{1e-200, 1e-200}, ErrUnderflow},
		{"negative zero from non-zero operands", negativeZero, []float64{-1e-200, 1e-200}, ErrUnderflow},
		{"zero from a zero first operand", 0, []float64{0, 5}, nil},
		{"zero from a zero second operand", 0, []float64{5, 0}, nil},
		{"zero from a single zero operand", 0, []float64{0}, nil},
		{"non-zero result", 2.5, []float64{10, 4}, nil},
		{"subnormal result", math.SmallestNonzeroFloat64, []float64{1e-300, 1e-24}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkUnderflow(tt.result, tt.operands); !errors.Is(got, tt.want) {
				t.Errorf("checkUnderflow(%v, %v) = %v, want %v", tt.result, tt.operands, got, tt.want)
			}
		})
	}
}

func TestGuardResult(t *testing.T) {
	underflowChecked := Operation{Name: "guarded", Arity: 2, CheckUnderflow: true}
	underflowExempt := Operation{Name: "exempt", Arity: 2, CheckUnderflow: false}

	tests := []struct {
		name     string
		op       Operation
		operands []float64
		result   float64
		want     error
	}{
		{"guarded operation underflows", underflowChecked, []float64{1e-200, 1e-200}, 0, ErrUnderflow},
		{"guarded operation with a zero operand", underflowChecked, []float64{0, 5}, 0, nil},
		{"exempt operation returns zero", underflowExempt, []float64{5, 5}, 0, nil},
		{"exempt operation returns negative zero", underflowExempt, []float64{5, 5}, negativeZero, nil},
		{"infinity outranks the underflow check", underflowChecked, []float64{1e308, 10}, math.Inf(1), ErrOverflow},
		{"not a number outranks the underflow check", underflowChecked, []float64{2, 3}, math.NaN(), ErrUndefined},
		{"ordinary result", underflowChecked, []float64{6, 7}, 42, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := guardResult(tt.op, tt.operands, tt.result); !errors.Is(got, tt.want) {
				t.Errorf("guardResult = %v, want %v", got, tt.want)
			}
		})
	}
}
