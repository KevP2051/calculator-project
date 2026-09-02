package calc

import "math"

func init() {
	register(Operation{Name: "add", Arity: 2, Apply: add})
	register(Operation{Name: "subtract", Arity: 2, Apply: subtract})
	register(Operation{Name: "multiply", Arity: 2, Apply: multiply, CheckUnderflow: true})
	register(Operation{Name: "divide", Arity: 2, Apply: divide, CheckUnderflow: true})
	register(Operation{Name: "power", Arity: 2, Apply: power, CheckUnderflow: true})
}

func Compute(op Operation, operands []float64) (float64, error) {
	result, err := op.Apply(operands)
	if err != nil {
		return 0, err
	}

	if err := guardResult(op, operands, result); err != nil {
		return 0, err
	}

	return result, nil
}

func add(operands []float64) (float64, error) {
	return operands[0] + operands[1], nil
}

func subtract(operands []float64) (float64, error) {
	return operands[0] - operands[1], nil
}

func multiply(operands []float64) (float64, error) {
	return operands[0] * operands[1], nil
}

func divide(operands []float64) (float64, error) {
	if operands[1] == 0 {
		return 0, ErrDivisionByZero
	}

	return operands[0] / operands[1], nil
}

func power(operands []float64) (float64, error) {
	base, exponent := operands[0], operands[1]

	if base < 0 && exponent != math.Trunc(exponent) {
		return 0, ErrOutOfDomain
	}

	return math.Pow(base, exponent), nil
}
