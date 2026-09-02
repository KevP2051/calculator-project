package api

import (
	"net/http"

	"calculator/backend/internal/service"
)

var statusByCode = map[service.Code]int{
	service.CodeMalformedJSON:        http.StatusBadRequest,
	service.CodeMissingField:         http.StatusBadRequest,
	service.CodeUnsupportedOperation: http.StatusBadRequest,
	service.CodeInvalidOperandCount:  http.StatusBadRequest,
	service.CodeInvalidOperand:       http.StatusBadRequest,
	service.CodeOperandOutOfRange:    http.StatusBadRequest,

	service.CodeDivisionByZero:     http.StatusUnprocessableEntity,
	service.CodeOperandOutOfDomain: http.StatusUnprocessableEntity,
	service.CodeResultOverflow:     http.StatusUnprocessableEntity,
	service.CodeResultUnderflow:    http.StatusUnprocessableEntity,
	service.CodeResultUndefined:    http.StatusUnprocessableEntity,
}
