export const OPERATIONS = [
    { value: "add", symbol: "+", label: "Addition", arity: 2 },
    { value: "subtract", symbol: "−", label: "Subtraction", arity: 2 },
    { value: "multiply", symbol: "×", label: "Multiplication", arity: 2 },
    { value: "divide", symbol: "÷", label: "Division", arity: 2 },
    { value: "power", symbol: "^", label: "Exponentiation", arity: 2 },
    { value: "sqrt", symbol: "√", label: "Square Root", arity: 1 },
    { value: "percentage", symbol: "%", label: "Percentage", arity: 2 },
] as const;

export type Operation =
    (typeof OPERATIONS)[number]["value"];

export const findOperation = (operation: Operation) =>
    OPERATIONS.find((entry) => entry.value === operation)!;

export const operationArity = (operation: Operation) => findOperation(operation).arity;
