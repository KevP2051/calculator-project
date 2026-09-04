
import { z } from 'zod'

export const calculationSchema = z.object({
    firstNumber: z.number({ error: 'Please enter a valid number' }),
    secondNumber: z.number({ error: 'Please enter a valid number' }),
    operation: z.enum(['add', 'subtract', 'multiply', 'divide'], { message: 'Invalid operation' }),
})

export type CalculatorFormData = z.infer<typeof calculationSchema>