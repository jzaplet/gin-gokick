package icu

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

const maxSafe = 1<<53 - 1

const maxDecimalDigits = 15

const floatDigits = 3

var ErrArgumentRange = errors.New("icu: argument out of range")

type Decimal struct {
	Value float64

	Digits int
}

type amount struct {
	value float64

	whole bool

	fewest int

	most int
}

type shown struct {
	integer int64

	fraction string

	negative bool
}

func amountOf(name string, arg any) (amount, error) {
	if decimal, ok := arg.(Decimal); ok {
		if decimal.Digits < 0 || decimal.Digits > maxDecimalDigits || math.IsNaN(decimal.Value) || math.IsInf(decimal.Value, 0) {
			return amount{}, fmt.Errorf("%w: %s is %v, a finite value with 0 to %d digits", ErrArgumentRange, name, decimal, maxDecimalDigits)
		}

		return amount{value: decimal.Value, fewest: decimal.Digits, most: decimal.Digits}, nil
	}

	value := reflect.ValueOf(arg)
	switch {
	case value.CanInt() && (value.Int() < -maxSafe || value.Int() > maxSafe):
		return amount{}, fmt.Errorf("%w: %s is %d, beyond 2^53", ErrArgumentRange, name, value.Int())
	case value.CanInt():
		return amount{value: float64(value.Int()), whole: true}, nil
	case value.Kind() == reflect.Float64 && (math.IsNaN(value.Float()) || math.IsInf(value.Float(), 0)):
		return amount{}, fmt.Errorf("%w: %s is %v", ErrArgumentRange, name, value.Float())
	case value.Kind() == reflect.Float64:
		return amount{value: value.Float(), most: floatDigits}, nil
	default:
		return amount{}, fmt.Errorf("%w: %s is a plural, not %T", ErrArgumentType, name, arg)
	}
}

func (a amount) exact() (int, bool) {
	if a.value != math.Trunc(a.value) || math.Abs(a.value) > maxSafe {
		return 0, false
	}

	return int(a.value), true
}

func (a amount) minus(offset int) (shown, error) {
	if a.whole == false {
		return round(a.value-float64(offset), a.fewest, a.most)
	}

	n := int64(a.value) - int64(offset)
	if n < -maxSafe || n > maxSafe {
		return shown{}, fmt.Errorf("%w: %d is beyond 2^53", ErrArgumentRange, n)
	}

	return shown{integer: max(n, -n), negative: n < 0}, nil
}

func round(value float64, fewest, most int) (shown, error) {
	whole, fraction, _ := strings.Cut(strconv.FormatFloat(math.Abs(value), 'f', -1, 64), ".")

	digits := []byte(whole + fraction[:min(len(fraction), most)])
	if len(fraction) > most && fraction[most] >= '5' {
		digits = increment(digits)
	}

	fraction = string(digits[len(digits)-min(len(fraction), most):])

	integer, err := strconv.ParseInt(string(digits[:len(digits)-len(fraction)]), 10, 64)
	if err != nil || integer > maxSafe {
		return shown{}, fmt.Errorf("%w: %v is beyond 2^53", ErrArgumentRange, value)
	}

	fraction = strings.TrimRight(fraction, "0")
	fraction += strings.Repeat("0", max(fewest-len(fraction), 0))

	return shown{integer: integer, fraction: fraction, negative: math.Signbit(value)}, nil
}

func increment(digits []byte) []byte {
	for i, digit := range slices.Backward(digits) {
		if digit < '9' {
			digits[i]++

			return digits
		}

		digits[i] = '0'
	}

	return append([]byte{'1'}, digits...)
}

func (s shown) operands() (i, v, w, f, t int) {
	trimmed := strings.TrimRight(s.fraction, "0")
	f, _ = strconv.Atoi("0" + s.fraction)
	t, _ = strconv.Atoi("0" + trimmed)

	return int(s.integer), len(s.fraction), len(trimmed), f, t
}
