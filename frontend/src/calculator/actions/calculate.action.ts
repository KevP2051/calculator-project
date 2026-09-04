import { calculatorApi } from "@/common/api/calculator.api"
import type { Operation } from "../constants/operations"


export const calculateAction = async (operation: Operation, operands: number[]) => {

    const { data } = await calculatorApi.post('/calculate', {
        operation,
        operands
    })

    return data;
}