package calc

import "errors"

var (
	ErrDivisionByZero = errors.New("division by zero")
	ErrOutOfDomain    = errors.New("operand outside the domain of the operation")
	ErrOverflow       = errors.New("result overflowed the representable range")
	ErrUnderflow      = errors.New("result underflowed to zero")
	ErrUndefined      = errors.New("result is undefined")
)
