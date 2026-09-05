package service

import (
	"errors"

	"calculator/backend/internal/calc"
)

type Code string

const (
	CodeMalformedJSON        Code = "MALFORMED_JSON"
	CodeMissingField         Code = "MISSING_FIELD"
	CodeUnsupportedOperation Code = "UNSUPPORTED_OPERATION"
	CodeInvalidOperandCount  Code = "INVALID_OPERAND_COUNT"
	CodeInvalidOperand       Code = "INVALID_OPERAND"
	CodeOperandOutOfRange    Code = "OPERAND_OUT_OF_RANGE"
	CodeDivisionByZero       Code = "DIVISION_BY_ZERO"
	CodeOperandOutOfDomain   Code = "OPERAND_OUT_OF_DOMAIN"
	CodeResultOverflow       Code = "RESULT_OVERFLOW"
	CodeResultUnderflow      Code = "RESULT_UNDERFLOW"
	CodeResultUndefined      Code = "RESULT_UNDEFINED"
)

var AllCodes = []Code{
	CodeMalformedJSON,
	CodeMissingField,
	CodeUnsupportedOperation,
	CodeInvalidOperandCount,
	CodeInvalidOperand,
	CodeOperandOutOfRange,
	CodeDivisionByZero,
	CodeOperandOutOfDomain,
	CodeResultOverflow,
	CodeResultUnderflow,
	CodeResultUndefined,
}

type Error struct {
	Code    Code
	Message string
}

var codeByCalcError = map[error]Code{
	calc.ErrDivisionByZero: CodeDivisionByZero,
	calc.ErrOutOfDomain:    CodeOperandOutOfDomain,
	calc.ErrOverflow:       CodeResultOverflow,
	calc.ErrUnderflow:      CodeResultUnderflow,
	calc.ErrUndefined:      CodeResultUndefined,
}

func errorForCalcFailure(err error) (*Error, bool) {
	for sentinel, code := range codeByCalcError {
		if errors.Is(err, sentinel) {
			return &Error{Code: code, Message: err.Error()}, true
		}
	}
	return nil, false
}
