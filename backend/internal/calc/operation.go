package calc

type Operation struct {
	Name           string
	Arity          int
	Apply          func(operands []float64) (float64, error)
	CheckUnderflow bool
}

var registry = map[string]Operation{}

func register(op Operation) {
	registry[op.Name] = op
}

func Lookup(name string) (Operation, bool) {
	op, ok := registry[name]
	return op, ok
}
