package icu

import (
	"math"
	"slices"
	"strconv"
	"strings"
)

const sampledIntegers = 1000

const sampledDecimalIntegers = 110

var largeIntegers = [...]int64{10_000, 100_000, 1_000_000, 2_000_000, 10_000_000}

func CategoryOrder() []string {
	return []string{"zero", "one", "two", "few", "many", other}
}

func (f *Formatter) Categories() []string {
	found := map[string]bool{}
	for i := range int64(sampledIntegers + 1) {
		found[f.category(shown{integer: i})] = true
	}

	for _, i := range largeIntegers {
		found[f.category(shown{integer: i})] = true
	}

	for digits := 1; digits <= floatDigits; digits++ {
		for fraction := range int(math.Pow10(digits)) {
			sample := shown{fraction: padded(fraction, digits)}
			for sample.integer = range int64(sampledDecimalIntegers + 1) {
				found[f.category(sample)] = true
			}
		}
	}

	return slices.DeleteFunc(CategoryOrder(), func(category string) bool { return found[category] == false })
}

func padded(n, digits int) string {
	s := strconv.Itoa(n)

	return strings.Repeat("0", digits-len(s)) + s
}
