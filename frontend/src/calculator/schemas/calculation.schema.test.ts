import { describe, expect, it } from 'vitest'
import { calculationSchema } from './calculation.schema'
import { OPERATIONS } from '../constants/operations'

const VALID_INPUT = { firstNumber: 25, secondNumber: 4, operation: 'add' }

const parse = (overrides: Record<string, unknown> = {}) =>
    calculationSchema.safeParse({ ...VALID_INPUT, ...overrides })

const messageFor = (result: ReturnType<typeof parse>, field: string) =>
    result.success
        ? undefined
        : result.error.issues.find((issue) => issue.path[0] === field)?.message

describe('calculationSchema', () => {

    describe('accepted input', () => {

        it('accepts two numbers and a supported operation', () => {
            expect(parse().success).toBe(true)
        })

        it('accepts negative operands', () => {
            expect(parse({ firstNumber: -25, secondNumber: -4.5 }).success).toBe(true)
        })

        it('accepts decimal operands', () => {
            expect(parse({ firstNumber: 0.1, secondNumber: 0.2 }).success).toBe(true)
        })

        it('accepts zero as an operand', () => {
            expect(parse({ firstNumber: 0, secondNumber: 0 }).success).toBe(true)
        })

        it.each(OPERATIONS.map((operation) => operation.value))(
            'accepts the %s operation offered by the UI',
            (operation) => {
                expect(parse({ operation }).success).toBe(true)
            }
        )
    })

    describe('rejected operands', () => {

        it('rejects NaN, which is what an empty number input produces', () => {
            const result = parse({ firstNumber: NaN })

            expect(result.success).toBe(false)
            expect(messageFor(result, 'firstNumber')).toBe('Please enter a valid number')
        })

        it('rejects non-finite operands the backend could not represent', () => {
            expect(parse({ firstNumber: Infinity }).success).toBe(false)
            expect(parse({ secondNumber: -Infinity }).success).toBe(false)
        })

        it('rejects a numeric string instead of coercing it', () => {
            expect(parse({ firstNumber: '25' }).success).toBe(false)
        })

        it('reports every missing operand rather than only the first', () => {
            const result = calculationSchema.safeParse({ operation: 'add' })

            expect(messageFor(result, 'firstNumber')).toBe('Please enter a valid number')
            expect(messageFor(result, 'secondNumber')).toBe('Please enter a valid number')
        })
    })

    describe('rejected operation', () => {

        it('rejects an operation the backend does not expose', () => {
            const result = parse({ operation: 'modulo' })

            expect(result.success).toBe(false)
            expect(messageFor(result, 'operation')).toBe('Invalid operation')
        })

        it('rejects a missing operation', () => {
            const result = calculationSchema.safeParse({ firstNumber: 25, secondNumber: 4 })

            expect(result.success).toBe(false)
        })
    })
})
