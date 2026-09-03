export const OPERATIONS = [
    { value: "add", symbol: "+", label: "Addition" },
    { value: "subtract", symbol: "−", label: "Subtraction" },
    { value: "multiply", symbol: "×", label: "Multiplication" },
    { value: "divide", symbol: "÷", label: "Division" },
    { value: "power", symbol: "^", label: "Exponentiation" },
    { value: "sqrt", symbol: "√", label: "Square Root" },
    { value: "percentage", symbol: "%", label: "Percentage" },
] as const;

export type Operation =
    (typeof OPERATIONS)[number]["value"];