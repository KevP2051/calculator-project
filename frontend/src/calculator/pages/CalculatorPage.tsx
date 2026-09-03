import { Field, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { OPERATIONS } from "../constants/operations"
import { Button } from "@/components/ui/button"

const numberFieldClassName =
    "h-12 rounded-2xl text-right text-lg font-medium tabular-nums [-moz-appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"


const CalculatorPage = () => {
    return (
        <div className="flex min-h-svh w-full items-center justify-center bg-background p-6">
            <div className="w-full max-w-sm rounded-4xl border border-border/60 bg-card p-8 shadow-xl shadow-foreground/5 ring-1 ring-foreground/5">

                <header className="mb-7 text-center">
                    <h1 className="text-lg font-semibold tracking-tight text-foreground">Calculator</h1>
                    <p className="mt-1 text-sm text-muted-foreground">Enter two numbers and pick an operation.</p>
                </header>

                <form className="flex flex-col gap-5" noValidate>

                    <Field>
                        <FieldLabel htmlFor="firstNumber">First number</FieldLabel>

                        <Input
                            id='firstNumber'
                            type="number"
                            inputMode="decimal"
                            placeholder="0"
                            className={numberFieldClassName}
                        />

                    </Field>

                    <Field>
                        <FieldLabel htmlFor="operation">
                            Operation
                        </FieldLabel>


                        <Select
                        >
                            <SelectTrigger id="operation" className="h-12 w-full rounded-2xl text-base">
                                <SelectValue />
                            </SelectTrigger>

                            <SelectContent>
                                {OPERATIONS.map((operation) => (
                                    <SelectItem
                                        key={operation.value}
                                        value={operation.value}
                                    >
                                        <span className="w-4 text-muted-foreground">{operation.symbol}</span>
                                        {operation.label}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>


                    </Field>

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
                        />


                    </Field>

                    <Button type="submit" size="lg" className="mt-1 h-12 rounded-2xl text-base font-semibold">
                        Calculate
                    </Button>

                    <div className="rounded-3xl bg-muted/50 px-5 py-4 text-center">
                        <p className="text-xs font-medium text-muted-foreground">Result</p>
                        <p className="mt-1 text-3xl font-semibold tabular-nums tracking-tight text-foreground">{'-'}</p>
                    </div>

                </form>
            </div>
        </div>
    )
}

export default CalculatorPage
