
import { useMutation } from '@tanstack/react-query'
import { isAxiosError } from 'axios';
import { toast } from 'sonner';
import type { Operation } from '../constants/operations';
import { calculateAction } from '../actions/calculate.action';
import {
    CALCULATION_ERROR_MESSAGES,
    NETWORK_ERROR_MESSAGE,
    UNKNOWN_ERROR_MESSAGE
} from '../constants/error-messages';

const resolveErrorMessage = (error: Error) => {

    if (!isAxiosError(error)) return UNKNOWN_ERROR_MESSAGE;

    if (!error.response) return NETWORK_ERROR_MESSAGE;

    const code = error.response.data?.error?.code;

    return CALCULATION_ERROR_MESSAGES[code] ?? UNKNOWN_ERROR_MESSAGE;
}

const useCalculate = () => {

    return useMutation({
        mutationFn: async ({ operation, operands }: { operation: Operation, operands: number[] }) =>
            calculateAction(operation, operands),
        retry: false,
        onError: (error) => {
            toast.error(resolveErrorMessage(error));
        }
    })
}

export default useCalculate
