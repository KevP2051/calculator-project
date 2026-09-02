package calc

import "math"

func checkFinite(result float64) error {
	switch {
	case math.IsNaN(result):
		return ErrUndefined
	case math.IsInf(result, 0):
		return ErrOverflow
	}
	return nil
}

func checkUnderflow(result float64, operands []float64) error {
	if result != 0 {
		return nil
	}
	for _, operand := range operands {
		if operand == 0 {
			return nil
		}
	}
	return ErrUnderflow
}

func guardResult(op Operation, operands []float64, result float64) error {
	if err := checkFinite(result); err != nil {
		return err
	}
	if op.CheckUnderflow {
		return checkUnderflow(result, operands)
	}
	return nil
}
