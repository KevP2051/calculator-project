import { Field, FieldError, FieldLabel } from "@/components/ui/field"
import { Controller, useForm, useWatch } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { OPERATIONS, findOperation, operationArity, type Operation } from "../constants/operations"
import { Button } from "@/components/ui/button"
import { calculationSchema, type CalculatorFormData } from "../schemas/calculation.schema"
import useCalculate from "../hooks/useCalculate"

const numberFieldClassName =
    "h-12 rounded-2xl text-right text-lg font-medium tabular-nums [-moz-appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"


const OperationLabel = ({ operation }: { operation: Operation }) => {

    const { symbol, label } = findOperation(operation)

    return (
        <>
            <span className="w-4 text-muted-foreground">{symbol}</span>
            {label}
        </>
    )
}


const CalculatorPage = () => {

    const calculatorForm = useForm<CalculatorFormData>({
        resolver: zodResolver(calculationSchema),
        defaultValues: {
            operation: "add"
        },
        mode: "onSubmit"
    })

    const calculate = useCalculate()

    const selectedOperation = useWatch({ control: calculatorForm.control, name: 'operation' })
    const takesSecondNumber = operationArity(selectedOperation) === 2

    const onSubmit = (formData: CalculatorFormData) => {
        calculate.mutate({
            operation: formData.operation,
            operands: operationArity(formData.operation) === 1
                ? [formData.firstNumber]
                : [formData.firstNumber, formData.secondNumber]
        })
    }

    return (
        <div className="flex min-h-svh w-full items-center justify-center bg-background p-6">
            <div className="w-full max-w-sm rounded-4xl border border-border/60 bg-card p-8 shadow-xl shadow-foreground/5 ring-1 ring-foreground/5">

                <header className="mb-7 text-center">
                    <h1 className="text-lg font-semibold tracking-tight text-foreground">Calculator</h1>
                    <p className="mt-1 text-sm text-muted-foreground">Enter two numbers and pick an operation.</p>
                </header>

                <form onSubmit={calculatorForm.handleSubmit(onSubmit)} className="flex flex-col gap-5" noValidate>

                    <Field>
                        <FieldLabel htmlFor="firstNumber">First number</FieldLabel>

                        <Input
                            id='firstNumber'
                            type="number"
                            inputMode="decimal"
                            placeholder="0"
                            className={numberFieldClassName}
                            {...calculatorForm.register('firstNumber', { valueAsNumber: true })}
                        />

                        {calculatorForm.formState.errors.firstNumber && (
                            <FieldError>
                                {calculatorForm.formState.errors.firstNumber.message}
                            </FieldError>
                        )}
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="operation">
                            Operation
                        </FieldLabel>

                        <Controller
                            name="operation"
                            control={calculatorForm.control}
                            render={({ field, fieldState }) => (
                                <>
                                    <Select
                                        value={field.value}
                                        onValueChange={(value) => {
                                            field.onChange(value)

                                            if (value && operationArity(value as Operation) === 1) {
                                                calculatorForm.setValue('secondNumber', 0, { shouldValidate: true })
                                            }
                                        }}
                                    >
                                        <SelectTrigger id="operation" className="h-12 w-full rounded-2xl text-base">
                                            <SelectValue className="gap-2">
                                                {(value: Operation) => <OperationLabel operation={value} />}
                                            </SelectValue>
                                        </SelectTrigger>

                                        <SelectContent>
                                            {OPERATIONS.map((operation) => (
                                                <SelectItem
                                                    key={operation.value}
                                                    value={operation.value}
                                                >
                                                    <OperationLabel operation={operation.value} />
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>

                                    {fieldState.error && (
                                        <FieldError>
                                            {fieldState.error.message}
                                        </FieldError>
                                    )}
                                </>
                            )}
                        />
                    </Field>

                    {takesSecondNumber && (
                        <Field>
                            <FieldLabel htmlFor="secondNumber">
                                Second number
                            </FieldLabel>

                            <Input
                                id="secondNumber"
                                type="number"
                                inputMode="decimal"
                                placeholder="0"
                                className={numberFieldClassName}
                                {...calculatorForm.register('secondNumber', { valueAsNumber: true })}
                            />

                            {calculatorForm.formState.errors.secondNumber && (
                                <FieldError>
                                    {calculatorForm.formState.errors.secondNumber.message}
                                </FieldError>
                            )}
                        </Field>
                    )}

                    <Button type="submit" size="lg" className="mt-1 h-12 rounded-2xl text-base font-semibold">
                        Calculate
                    </Button>

                    <div className="rounded-3xl bg-muted/50 px-5 py-4 text-center">
                        <p className="text-xs font-medium text-muted-foreground">Result</p>
                        <p className="mt-1 text-3xl font-semibold tabular-nums tracking-tight text-foreground">{calculate.data?.result ?? '--'}</p>
                    </div>


                </form>
            </div>
        </div>
    )
}

export default CalculatorPage
