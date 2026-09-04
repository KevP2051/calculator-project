import CalculatorPage from "./calculator/pages/CalculatorPage"
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Toaster } from '@/components/ui/sonner'

const queryClient = new QueryClient()


const CalculatorApp = () => {
    return (
        <>
            <QueryClientProvider client={queryClient}>
                <CalculatorPage />
                <Toaster visibleToasts={1} position="top-center" richColors />
            </QueryClientProvider>

        </>
    )
}

export default CalculatorApp
