export const CALCULATION_ERROR_MESSAGES: Record<string, string> = {
    MALFORMED_JSON: "The request could not be read. Please try again.",
    MISSING_FIELD: "Please complete every field before calculating.",
    UNSUPPORTED_OPERATION: "That operation is not available.",
    INVALID_OPERAND_COUNT: "That operation expects a different amount of numbers.",
    INVALID_OPERAND: "Please enter valid numbers.",
    OPERAND_OUT_OF_RANGE: "That number is too large to work with.",
    DIVISION_BY_ZERO: "Cannot divide by zero.",
    OPERAND_OUT_OF_DOMAIN: "That operation is not defined for those numbers.",
    RESULT_OVERFLOW: "The result is too large to represent.",
    RESULT_UNDERFLOW: "The result is too small to represent.",
    RESULT_UNDEFINED: "The result is not a number.",
};

export const NETWORK_ERROR_MESSAGE = "Cannot reach the server. Please try again.";

export const UNKNOWN_ERROR_MESSAGE = "Something went wrong. Please try again.";
