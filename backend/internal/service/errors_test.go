package service

import (
	"errors"
	"testing"

	"calculator/backend/internal/calc"
)

func TestErrorForCalcFailureMapsEverySentinel(t *testing.T) {
	tests := []struct {
		name     string
		sentinel error
		want     Code
	}{
		{"division by zero", calc.ErrDivisionByZero, CodeDivisionByZero},
		{"out of domain", calc.ErrOutOfDomain, CodeOperandOutOfDomain},
		{"overflow", calc.ErrOverflow, CodeResultOverflow},
		{"underflow", calc.ErrUnderflow, CodeResultUnderflow},
		{"undefined", calc.ErrUndefined, CodeResultUndefined},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := errorForCalcFailure(tt.sentinel)
			if !ok {
				t.Fatalf("errorForCalcFailure(%v) reported no mapping", tt.sentinel)
			}
			if got.Code != tt.want {
				t.Errorf("code = %q, want %q", got.Code, tt.want)
			}
			if got.Message == "" {
				t.Error("message is empty")
			}
		})
	}
}

func TestErrorForCalcFailureUnwrapsWrappedSentinels(t *testing.T) {
	wrapped := errors.Join(errors.New("context"), calc.ErrOverflow)

	got, ok := errorForCalcFailure(wrapped)
	if !ok {
		t.Fatal("errorForCalcFailure reported no mapping for a wrapped sentinel")
	}
	if got.Code != CodeResultOverflow {
		t.Errorf("code = %q, want %q", got.Code, CodeResultOverflow)
	}
}

func TestErrorForCalcFailureRejectsUnknownError(t *testing.T) {
	if _, ok := errorForCalcFailure(errors.New("something else")); ok {
		t.Error("errorForCalcFailure claimed a mapping for an unknown error")
	}
}

func TestAllCodesAreUniqueAndNonEmpty(t *testing.T) {
	seen := make(map[Code]bool, len(AllCodes))

	for _, code := range AllCodes {
		if code == "" {
			t.Error("AllCodes contains an empty code")
		}
		if seen[code] {
			t.Errorf("code %q appears more than once", code)
		}
		seen[code] = true
	}

	if len(AllCodes) != 11 {
		t.Errorf("len(AllCodes) = %d, want 11", len(AllCodes))
	}
}
