
import { useMutation } from '@tanstack/react-query'
import type { Operation } from '../constants/operations';
import { calculateAction } from '../actions/calculate.action';

const useCalculate = () => {

    return useMutation({
        mutationFn: async ({ operation, operands }: { operation: Operation, operands: number[] }) =>
            calculateAction(operation, operands),
        retry: false,
        onSuccess: (data) => {
            console.log('Calculation successful:', data);
        },
        onError: (error) => {
            console.error('Error occurred while calculating:', error);
        }


    })
}

export default useCalculate
