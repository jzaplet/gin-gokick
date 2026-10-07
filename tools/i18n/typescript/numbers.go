package typescript

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"gokick/app/core/i18n/icu"
	"gokick/app/core/locale/posix"
	"gokick/locale"
	"gokick/tools/codegen"
	"gokick/tools/i18n/dictionaries"
)

const NumbersDir = "tests/assets/i18n/numbers"

const smallCounts = 26

var counts = []int{100, 101, 102, 103, 105, 111, 112, 1000, 1001, 1234, 10_000, 12_345, 100_000, 1_000_000, 1_234_567, 2_000_000, 10_000_000, 1<<53 - 1, -1, -2, -5, -1234, -(1<<53 - 1)}

var fractions = []float64{0.1, 0.5, 1.25, 1.5, 2.5, 21.1, 1234.5, 1_234_567.891, 1.23456, 0.0005, 0.9995, 1e-7, -2.5, -0.0004, math.Copysign(0, -1)}

var decimals = []icu.Decimal{
	{Value: 1, Digits: 1},
	{Value: 2, Digits: 1},
	{Value: 5, Digits: 1},
	{Value: 19.9, Digits: 2},
	{Value: 0.125, Digits: 2},
	{Value: 1.005, Digits: 2},
	{Value: 2.5, Digits: 0},
	{Value: 1234.5678, Digits: 2},
	{Value: -1.5, Digits: 0},
	{Value: -0.0004, Digits: 2},
	{Value: 1, Digits: 15},
}

type sample struct {
	arg any

	literal string
}

func Numbers(c dictionaries.Catalog) (codegen.Files, error) {
	m, err := icu.Parse(everyCategory())
	if err != nil {
		return codegen.Files{}, err
	}

	content := map[string][]byte{NumbersDir + "/sample.ts": sampleModule(m)}
	for _, l := range c.Locales {
		module, err := shownNumbers(l.Name, m)
		if err != nil {
			return codegen.Files{}, err
		}

		content[NumbersDir+"/"+l.Name+".ts"] = module
	}

	return codegen.Files{Header: Header, Dir: NumbersDir, Content: content}, nil
}

func everyCategory() string {
	var b strings.Builder
	b.WriteString("{n, plural,")

	for _, category := range icu.CategoryOrder() {
		fmt.Fprintf(&b, " %s {%s #}", category, category)
	}

	b.WriteString("}")

	return b.String()
}

func sampleModule(m icu.Message) []byte {
	var b strings.Builder
	b.WriteString(Header + "\n")
	b.WriteString("import type { Decimal } from '" + importPath(locale.TypeScriptDir+"/types/Decimal.ts") + "';\n")
	b.WriteString("import type { Message } from '" + importPath(locale.TypeScriptDir+"/types/Message.ts") + "';\n\n")
	b.WriteString("export type NumberSample = readonly [Decimal | number, string];\n\n")
	fmt.Fprintf(&b, "export const message: Message = %s;\n", message(m))

	return []byte(b.String())
}

func shownNumbers(name string, m icu.Message) ([]byte, error) {
	parsed, err := posix.Parse(name)
	if err != nil {
		return nil, err
	}

	f := icu.NewFormatter(parsed.Tag())
	var b strings.Builder
	b.WriteString(Header + "\n")
	b.WriteString("import type { NumberSample } from './sample';\n\n")
	fmt.Fprintf(&b, "export const tag = %s;\n\n", codegen.Quote(strings.Replace(name, "_", "-", 1)))
	b.WriteString("export const samples: readonly NumberSample[] = [\n")

	for _, s := range samples() {
		text, err := f.Format(m, map[string]any{"n": s.arg})
		if err != nil {
			return nil, fmt.Errorf("%s %s: %w", name, s.literal, err)
		}

		fmt.Fprintf(&b, "    [%s, %s],\n", s.literal, codegen.Quote(text))
	}

	b.WriteString("];\n")

	return []byte(b.String()), nil
}

func samples() []sample {
	all := make([]sample, 0, smallCounts+len(counts)+len(fractions)+len(decimals))
	for n := range smallCounts {
		all = append(all, sample{arg: n, literal: strconv.Itoa(n)})
	}

	for _, n := range counts {
		all = append(all, sample{arg: n, literal: strconv.Itoa(n)})
	}

	for _, n := range fractions {
		all = append(all, sample{arg: n, literal: float(n)})
	}

	for _, d := range decimals {
		all = append(all, sample{arg: d, literal: fmt.Sprintf("{ value: %s, digits: %d }", float(d.Value), d.Digits)})
	}

	return all
}

func float(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}
