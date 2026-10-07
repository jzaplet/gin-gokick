package icu

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

var ErrMissingArgument = errors.New("icu: missing argument")

var ErrUnknownArgument = errors.New("icu: unknown argument")

var ErrArgumentType = errors.New("icu: argument of the wrong type")

var categories = [...]string{
	plural.Other: other,

	plural.Zero: "zero",

	plural.One: "one",

	plural.Two: "two",

	plural.Few: "few",

	plural.Many: "many",
}

type Formatter struct {
	tag language.Tag

	printer *message.Printer

	minusBefore string

	minusAfter string

	separator string

	digits [10]string
}

type values struct {
	texts map[string]string

	amounts map[string]amount
}

func NewFormatter(tag language.Tag) *Formatter {
	printer := message.NewPrinter(tag)

	f := &Formatter{tag: tag, printer: printer}
	for digit := range f.digits {
		f.digits[digit] = printer.Sprint(number.Decimal(digit))
	}

	f.minusBefore, f.minusAfter, _ = strings.Cut(printer.Sprint(number.Decimal(-1)), f.digits[1])
	half := printer.Sprint(number.Decimal(0.5, number.MinFractionDigits(1), number.MaxFractionDigits(1)))
	f.separator = strings.TrimSuffix(strings.TrimPrefix(half, f.digits[0]), f.digits[5])

	return f
}

func (f *Formatter) Format(m Message, args map[string]any) (string, error) {
	vals, err := read(m, args)
	if err != nil {
		return "", err
	}

	var out strings.Builder
	if err := f.write(&out, m, vals, shown{}); err != nil {
		return "", err
	}

	return out.String(), nil
}

func (f *Formatter) write(out *strings.Builder, m Message, vals values, pound shown) error {
	for _, part := range m {
		var err error

		switch part := part.(type) {
		case Text:
			out.WriteString(string(part))
		case Arg:
			out.WriteString(vals.texts[part.Name])
		case Pound:
			out.WriteString(f.number(pound))
		case Plural:
			err = f.writePlural(out, part, vals)
		case Select:
			err = f.write(out, selectCase(part, vals.texts[part.Name]), vals, pound)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Formatter) writePlural(out *strings.Builder, part Plural, vals values) error {
	value := vals.amounts[part.Name]

	pound, err := value.minus(part.Offset)
	if err != nil {
		return fmt.Errorf("%s: %w", part.Name, err)
	}

	if exact, ok := value.exact(); ok {
		if m, ok := part.Exact[exact]; ok {
			return f.write(out, m, vals, pound)
		}
	}

	m, ok := part.Categories[f.category(pound)]
	if ok == false {
		m = part.Categories[other]
	}

	return f.write(out, m, vals, pound)
}

func (f *Formatter) category(s shown) string {
	i, v, w, fraction, t := s.operands()

	return categories[plural.Cardinal.MatchPlural(f.tag, i, v, w, fraction, t)]
}

func (f *Formatter) number(s shown) string {
	var out strings.Builder
	if s.negative {
		out.WriteString(f.minusBefore)
	}

	out.WriteString(f.printer.Sprint(number.Decimal(s.integer)))

	if s.fraction != "" {
		out.WriteString(f.separator)
	}

	for _, digit := range s.fraction {
		out.WriteString(f.digits[digit-'0'])
	}

	if s.negative {
		out.WriteString(f.minusAfter)
	}

	return out.String()
}

func selectCase(part Select, value string) Message {
	if m, ok := part.Cases[value]; ok {
		return m
	}

	return part.Cases[other]
}

func read(m Message, args map[string]any) (values, error) {
	kinds, err := m.Args()
	if err != nil {
		return values{}, err
	}

	for name := range args {
		if _, ok := kinds[name]; ok == false {
			return values{}, fmt.Errorf("%w: %s", ErrUnknownArgument, name)
		}
	}

	vals := values{texts: map[string]string{}, amounts: map[string]amount{}}

	for name, kind := range kinds {
		arg, ok := args[name]
		if ok == false {
			return values{}, fmt.Errorf("%w: %s", ErrMissingArgument, name)
		}

		if err := vals.add(name, kind, arg); err != nil {
			return values{}, err
		}
	}

	return vals, nil
}

func (v values) add(name string, kind Kind, arg any) error {
	if kind == KindPlural {
		value, err := amountOf(name, arg)
		v.amounts[name] = value

		return err
	}

	value := reflect.ValueOf(arg)
	if value.Kind() != reflect.String {
		return fmt.Errorf("%w: %s is %s, not %T", ErrArgumentType, name, kind, arg)
	}

	v.texts[name] = value.String()

	return nil
}
