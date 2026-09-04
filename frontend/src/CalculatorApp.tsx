import CalculatorPage from "./calculator/pages/CalculatorPage"
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

const queryClient = new QueryClient()


const CalculatorApp = () => {
    return (
        <>
            <QueryClientProvider client={queryClient}>
                <CalculatorPage />
            </QueryClientProvider>

        </>
    )
}

export default CalculatorApp