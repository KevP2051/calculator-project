import { describe, expect, it } from 'vitest'
import { formatResult } from './format-result'

describe('formatResult', () => {

    describe('floating point noise', () => {

        it.each([
            ['0.1 + 0.2', 0.1 + 0.2, '0.3'],
            ['1.1 + 2.2', 1.1 + 2.2, '3.3'],
            ['4.35 * 100', 4.35 * 100, '435'],
            ['0.3 - 0.1', 0.3 - 0.1, '0.2'],
        ])('hides the binary64 noise the backend returns for %s', (_operation, result, expected) => {
            expect(formatResult(result)).toBe(expected)
        })
    })

    describe('exactly representable results', () => {

        it('keeps every digit of an exact integer below 2^53', () => {
            expect(formatResult(999999999999 + 99999999999999)).toBe('100,999,999,999,998')
        })

        it.each([
            [29, '29'],
            [0, '0'],
            [-42.5, '-42.5'],
            [1234567.891, '1,234,567.891'],
        ])('returns %d unchanged', (result, expected) => {
            expect(formatResult(result)).toBe(expected)
        })

        it('groups thousands so long integers stay readable', () => {
            expect(formatResult(1234567)).toBe('1,234,567')
        })
    })

    describe('irrational results', () => {

        it.each([
            [Math.sqrt(2), '1.4142135623731'],
            [1 / 3, '0.333333333333333'],
        ])('caps %d at the fifteen digits binary64 can round-trip', (result, expected) => {
            expect(formatResult(result)).toBe(expected)
        })
    })

    describe('notation', () => {

        it('stays decimal at the largest value below the scientific threshold', () => {
            expect(formatResult(999999999999999)).toBe('999,999,999,999,999')
        })

        it('switches to scientific rather than printing hundreds of digits', () => {
            expect(formatResult(1e15)).toBe('1E15')
            expect(formatResult(1.0000000000000001e305)).toBe('1E305')
        })

        it('stays decimal at the smallest value above the lower threshold', () => {
            expect(formatResult(0.000001)).toBe('0.000001')
        })

        it('switches to scientific rather than printing a run of leading zeros', () => {
            expect(formatResult(1e-7)).toBe('1E-7')
            expect(formatResult(3.333333333333333e-301)).toBe('3.33333333333333E-301')
        })

        it('keeps zero in decimal notation even though it is below the lower threshold', () => {
            expect(formatResult(0)).toBe('0')
        })
    })

    describe('output length', () => {

        it.each([
            [999999999999999, 'largest decimal'],
            [-999999999999999, 'largest negative decimal'],
            [1.23456789012345e-301, 'longest scientific'],
            [-1.23456789012345e-301, 'longest negative scientific'],
        ])('keeps the %d case within what the result panel can show', (result) => {
            expect(formatResult(result).length).toBeLessThanOrEqual(22)
        })
    })
})
