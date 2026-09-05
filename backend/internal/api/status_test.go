package api

import (
	"net/http"
	"testing"

	"calculator/backend/internal/service"
)

func TestEveryCodeHasAStatus(t *testing.T) {
	for _, code := range service.AllCodes {
		status, ok := statusByCode[code]
		if !ok {
			t.Errorf("code %q has no status mapping", code)
			continue
		}
		if status == 0 {
			t.Errorf("code %q maps to status 0", code)
		}
	}
}

func TestStatusTableHasNoUnknownCodes(t *testing.T) {
	known := make(map[service.Code]bool, len(service.AllCodes))
	for _, code := range service.AllCodes {
		known[code] = true
	}

	for code := range statusByCode {
		if !known[code] {
			t.Errorf("status table contains unknown code %q", code)
		}
	}
}

func TestRequestFaultsAreBadRequest(t *testing.T) {
	requestFaults := []service.Code{
		service.CodeMalformedJSON,
		service.CodeMissingField,
		service.CodeUnsupportedOperation,
		service.CodeInvalidOperandCount,
		service.CodeInvalidOperand,
		service.CodeOperandOutOfRange,
	}

	for _, code := range requestFaults {
		if got := statusByCode[code]; got != http.StatusBadRequest {
			t.Errorf("status for %q = %d, want %d", code, got, http.StatusBadRequest)
		}
	}
}

func TestCalculationFaultsAreUnprocessableEntity(t *testing.T) {
	calculationFaults := []service.Code{
		service.CodeDivisionByZero,
		service.CodeOperandOutOfDomain,
		service.CodeResultOverflow,
		service.CodeResultUnderflow,
		service.CodeResultUndefined,
	}

	for _, code := range calculationFaults {
		if got := statusByCode[code]; got != http.StatusUnprocessableEntity {
			t.Errorf("status for %q = %d, want %d", code, got, http.StatusUnprocessableEntity)
		}
	}
}
