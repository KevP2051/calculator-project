package calc

import "testing"

func TestLookupReportsUnregisteredNames(t *testing.T) {
	if _, ok := Lookup("no_such_operation"); ok {
		t.Error("Lookup claimed to find an unregistered operation")
	}
}

func TestRegisterMakesOperationLookupable(t *testing.T) {
	const name = "operation_test_stub"
	t.Cleanup(func() { delete(registry, name) })

	register(Operation{
		Name:           name,
		Arity:          2,
		Apply:          func(operands []float64) (float64, error) { return operands[0], nil },
		CheckUnderflow: true,
	})

	op, ok := Lookup(name)
	if !ok {
		t.Fatal("Lookup did not find the registered operation")
	}
	if op.Name != name {
		t.Errorf("Name = %q, want %q", op.Name, name)
	}
	if op.Arity != 2 {
		t.Errorf("Arity = %d, want 2", op.Arity)
	}
	if !op.CheckUnderflow {
		t.Error("CheckUnderflow = false, want true")
	}
	if op.Apply == nil {
		t.Fatal("Apply is nil")
	}

	got, err := op.Apply([]float64{7, 9})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if got != 7 {
		t.Errorf("Apply = %v, want 7", got)
	}
}
