
import { z } from 'zod'
import { OPERATIONS } from '../constants/operations'

const operationValues = OPERATIONS.map((operation) => operation.value)

export const calculationSchema = z.object({
    firstNumber: z.number({ error: 'Please enter a valid number' }),
    secondNumber: z.number({ error: 'Please enter a valid number' }),
    operation: z.enum(operationValues, { message: 'Invalid operation' }),
})

export type CalculatorFormData = z.infer<typeof calculationSchema>
