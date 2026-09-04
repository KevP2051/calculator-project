import { calculatorApi } from "@/common/api/calculator.api"
import type { Operation } from "../constants/operations"
import type { CalculationRequest, CalculationResponse } from "../types/calculation.types"


export const calculateAction = async (operation: Operation, operands: number[]) => {

    const payload: CalculationRequest = { operation, operands }

    const { data } = await calculatorApi.post<CalculationResponse>('/calculate', payload)

    return data;
}
