import type { Operation } from '../constants/operations'

export interface CalculationRequest {
    operation: Operation
    operands: number[]
}

export interface CalculationResponse {
    result: number
}

export interface CalculationErrorResponse {
    error: {
        code: string
        message: string
    }
}
