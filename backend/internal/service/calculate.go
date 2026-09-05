package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"calculator/backend/internal/calc"
)

type Request struct {
	Operation string
	Operands  []any
}

func Calculate(req Request) (float64, *Error) {
	if req.Operation == "" {
		return 0, &Error{Code: CodeMissingField, Message: "operation is required"}
	}

	if req.Operands == nil {
		return 0, &Error{Code: CodeMissingField, Message: "operands is required"}
	}

	op, ok := calc.Lookup(req.Operation)
	if !ok {
		return 0, &Error{
			Code:    CodeUnsupportedOperation,
			Message: fmt.Sprintf("unsupported operation %q", req.Operation),
		}
	}

	if len(req.Operands) != op.Arity {
		return 0, &Error{
			Code:    CodeInvalidOperandCount,
			Message: fmt.Sprintf("%s takes %d operands, got %d", op.Name, op.Arity, len(req.Operands)),
		}
	}

	operands, operandErr := parseOperands(req.Operands)
	if operandErr != nil {
		return 0, operandErr
	}

	result, err := calc.Compute(op, operands)
	if err != nil {
		if failure, mapped := errorForCalcFailure(err); mapped {
			return 0, failure
		}
		return 0, &Error{Code: CodeResultUndefined, Message: err.Error()}
	}

	return result, nil
}

func parseOperands(values []any) ([]float64, *Error) {
	operands := make([]float64, len(values))

	for i, value := range values {
		number, ok := value.(json.Number)
		if !ok {
			return nil, &Error{
				Code:    CodeInvalidOperand,
				Message: fmt.Sprintf("operand %d is not a number", i),
			}
		}

		parsed, err := strconv.ParseFloat(number.String(), 64)
		switch {
		case errors.Is(err, strconv.ErrRange):
			return nil, &Error{
				Code:    CodeOperandOutOfRange,
				Message: fmt.Sprintf("operand %d is outside the range representable as a finite float64", i),
			}
		case err != nil:
			return nil, &Error{
				Code:    CodeInvalidOperand,
				Message: fmt.Sprintf("operand %d is not a number", i),
			}
		}

		operands[i] = parsed
	}

	return operands, nil
}
