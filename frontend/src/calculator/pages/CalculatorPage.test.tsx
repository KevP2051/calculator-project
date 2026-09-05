import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AxiosError, AxiosHeaders, type AxiosResponse } from 'axios'
import { toast } from 'sonner'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CalculatorPage from './CalculatorPage'
import { calculatorApi } from '../api/calculator.api'
import {
    CALCULATION_ERROR_MESSAGES,
    NETWORK_ERROR_MESSAGE,
    UNKNOWN_ERROR_MESSAGE,
} from '../constants/error-messages'

vi.mock('../api/calculator.api', () => ({
    calculatorApi: { post: vi.fn() },
}))

vi.mock('sonner', () => ({
    toast: { error: vi.fn() },
}))

const postMock = vi.mocked(calculatorApi.post)
const toastErrorMock = vi.mocked(toast.error)

const AXIOS_CONFIG = { headers: new AxiosHeaders() }

const backendReturns = (result: number) => {
    postMock.mockResolvedValue({ data: { result } })
}

const backendRejectsWith = (code: string) => {
    const response = {
        data: { error: { code, message: `backend detail for ${code}` } },
        status: 400,
        statusText: 'Bad Request',
        headers: new AxiosHeaders(),
        config: AXIOS_CONFIG,
    } as AxiosResponse

    postMock.mockRejectedValue(
        new AxiosError('Request failed', 'ERR_BAD_REQUEST', AXIOS_CONFIG, {}, response)
    )
}

const backendIsUnreachable = () => {
    postMock.mockRejectedValue(new AxiosError('Network Error', 'ERR_NETWORK', AXIOS_CONFIG, {}))
}

const backendHangs = () => {
    let release: (result: number) => void = () => {}

    postMock.mockReturnValue(new Promise((resolve) => {
        release = (result) => resolve({ data: { result } })
    }))

    return (result: number) => release(result)
}

const renderCalculatorPage = () => {
    const queryClient = new QueryClient({
        defaultOptions: { mutations: { retry: false } },
    })

    render(
        <QueryClientProvider client={queryClient}>
            <CalculatorPage />
        </QueryClientProvider>
    )

    return userEvent.setup()
}

const firstNumberInput = () => screen.getByLabelText('First number')
const secondNumberInput = () => screen.getByLabelText('Second number')
const calculateButton = () => screen.getByRole('button', { name: 'Calculate' })

const chooseOperation = async (user: ReturnType<typeof userEvent.setup>, label: string) => {
    await user.click(screen.getByLabelText('Operation'))
    await user.click(await screen.findByRole('option', { name: new RegExp(label) }))
}

describe('CalculatorPage', () => {

    beforeEach(() => {
        vi.clearAllMocks()
    })

    describe('initial render', () => {

        it('offers two operand inputs, an operation selector and a submit button', () => {
            renderCalculatorPage()

            expect(firstNumberInput()).toBeInTheDocument()
            expect(secondNumberInput()).toBeInTheDocument()
            expect(screen.getByLabelText('Operation')).toBeInTheDocument()
            expect(calculateButton()).toBeInTheDocument()
        })

        it('starts on addition with an empty result placeholder', () => {
            renderCalculatorPage()

            expect(screen.getByLabelText('Operation')).toHaveTextContent('Addition')
            expect(screen.getByText('--')).toBeInTheDocument()
        })
    })

    describe('entering a calculation', () => {

        it('keeps what the user types in each operand input', async () => {
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '25')
            await user.type(secondNumberInput(), '4')

            expect(firstNumberInput()).toHaveValue(25)
            expect(secondNumberInput()).toHaveValue(4)
        })

        it('lets the user pick a different operation', async () => {
            const user = renderCalculatorPage()

            await chooseOperation(user, 'Division')

            expect(screen.getByLabelText('Operation')).toHaveTextContent('Division')
        })

        it('asks only for one operand when the operation is unary', async () => {
            const user = renderCalculatorPage()

            await chooseOperation(user, 'Square Root')

            expect(firstNumberInput()).toBeInTheDocument()
            expect(screen.queryByLabelText('Second number')).not.toBeInTheDocument()
        })
    })

    describe('submitting to the backend', () => {

        it('sends the operands and the selected operation', async () => {
            backendReturns(6.25)
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '25')
            await user.type(secondNumberInput(), '4')
            await chooseOperation(user, 'Division')
            await user.click(calculateButton())

            await waitFor(() => expect(postMock).toHaveBeenCalledWith('/calculate', {
                operation: 'divide',
                operands: [25, 4],
            }))
        })

        it('sends a single operand for a unary operation', async () => {
            backendReturns(5)
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '25')
            await chooseOperation(user, 'Square Root')
            await user.click(calculateButton())

            await waitFor(() => expect(postMock).toHaveBeenCalledWith('/calculate', {
                operation: 'sqrt',
                operands: [25],
            }))
        })

        it('blocks a second submission while the first one is still in flight', async () => {
            const finishRequest = backendHangs()
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '6')
            await user.type(secondNumberInput(), '7')
            await user.click(calculateButton())

            expect(await screen.findByRole('button', { name: 'Calculating…' })).toBeDisabled()
            expect(postMock).toHaveBeenCalledTimes(1)

            finishRequest(42)

            await screen.findByText('42')
            expect(calculateButton()).toBeEnabled()
        })

        it('displays the result returned by the backend', async () => {
            backendReturns(29)
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '25')
            await user.type(secondNumberInput(), '4')
            await user.click(calculateButton())

            expect(await screen.findByText('29')).toBeInTheDocument()
        })

        it('displays a long result in full without truncating it', async () => {
            backendReturns(100999999999998)
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '999999999999')
            await user.type(secondNumberInput(), '99999999999999')
            await user.click(calculateButton())

            expect(await screen.findByText('100,999,999,999,998')).toBeInTheDocument()
        })

        it('displays the result formatted rather than raw', async () => {
            backendReturns(0.1 + 0.2)
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '0.1')
            await user.type(secondNumberInput(), '0.2')
            await user.click(calculateButton())

            expect(await screen.findByText('0.3')).toBeInTheDocument()
        })
    })

    describe('invalid input', () => {

        it('reports both operands and never reaches the backend when they are empty', async () => {
            const user = renderCalculatorPage()

            await user.click(calculateButton())

            expect(await screen.findAllByText('Please enter a valid number')).toHaveLength(2)
            expect(postMock).not.toHaveBeenCalled()
        })

        it('reports only the missing operand when the other one is filled', async () => {
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '25')
            await user.click(calculateButton())

            expect(await screen.findAllByText('Please enter a valid number')).toHaveLength(1)
            expect(postMock).not.toHaveBeenCalled()
        })
    })

    describe('backend errors', () => {

        it('shows the message mapped from the backend error code', async () => {
            backendRejectsWith('DIVISION_BY_ZERO')
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '6')
            await user.type(secondNumberInput(), '0')
            await chooseOperation(user, 'Division')
            await user.click(calculateButton())

            await waitFor(() => expect(toastErrorMock)
                .toHaveBeenCalledWith(CALCULATION_ERROR_MESSAGES.DIVISION_BY_ZERO))
        })

        it('never leaks the raw backend detail to the user', async () => {
            backendRejectsWith('DIVISION_BY_ZERO')
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '6')
            await user.type(secondNumberInput(), '0')
            await user.click(calculateButton())

            await waitFor(() => expect(toastErrorMock).toHaveBeenCalled())
            expect(toastErrorMock).not.toHaveBeenCalledWith(expect.stringContaining('backend detail'))
        })

        it('still tells the user something when the backend sends an unknown code', async () => {
            backendRejectsWith('A_CODE_THIS_FRONTEND_DOES_NOT_KNOW')
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '25')
            await user.type(secondNumberInput(), '4')
            await user.click(calculateButton())

            await waitFor(() => expect(toastErrorMock).toHaveBeenCalledWith(UNKNOWN_ERROR_MESSAGE))
        })

        it('still tells the user something when the failure is not an http error', async () => {
            postMock.mockRejectedValue(new Error('unexpected client failure'))
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '25')
            await user.type(secondNumberInput(), '4')
            await user.click(calculateButton())

            await waitFor(() => expect(toastErrorMock).toHaveBeenCalledWith(UNKNOWN_ERROR_MESSAGE))
        })

        it('reports a connection problem when the request never reaches the backend', async () => {
            backendIsUnreachable()
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '25')
            await user.type(secondNumberInput(), '4')
            await user.click(calculateButton())

            await waitFor(() => expect(toastErrorMock).toHaveBeenCalledWith(NETWORK_ERROR_MESSAGE))
        })

        it('leaves the result placeholder untouched when the calculation fails', async () => {
            backendRejectsWith('DIVISION_BY_ZERO')
            const user = renderCalculatorPage()

            await user.type(firstNumberInput(), '6')
            await user.type(secondNumberInput(), '0')
            await user.click(calculateButton())

            await waitFor(() => expect(toastErrorMock).toHaveBeenCalled())
            expect(screen.getByText('--')).toBeInTheDocument()
        })
    })
})
