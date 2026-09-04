const SIGNIFICANT_DIGITS = 15

const SCIENTIFIC_UPPER_BOUND = 1e15
const SCIENTIFIC_LOWER_BOUND = 1e-6

const decimalFormatter = new Intl.NumberFormat("en-US", {
    maximumSignificantDigits: SIGNIFICANT_DIGITS,
})

const scientificFormatter = new Intl.NumberFormat("en-US", {
    notation: "scientific",
    maximumSignificantDigits: SIGNIFICANT_DIGITS,
})

export const formatResult = (result: number) => {

    const magnitude = Math.abs(result)

    const needsScientificNotation =
        magnitude !== 0 &&
        (magnitude >= SCIENTIFIC_UPPER_BOUND || magnitude < SCIENTIFIC_LOWER_BOUND)

    return needsScientificNotation
        ? scientificFormatter.format(result)
        : decimalFormatter.format(result)
}
