package main

import (
	"errors"
	"flag"
	"fmt"
	"maps"
	"os"
	"reflect"
	"slices"

	"gokick/app/core/api"
	"gokick/locale"
	"gokick/tools/codegen"
	"gokick/tools/i18n/dictionaries"
	"gokick/tools/i18n/frontend"
	"gokick/tools/i18n/gocode"
	"gokick/tools/i18n/templates/keys"
	"gokick/tools/i18n/typescript"
	"gokick/tools/i18n/unused"
	"gokick/views"
)

func main() {
	os.Exit(run())
}

func run() int {
	check := flag.Bool("check", false, "write nothing, fail when a generated file differs from the dictionaries")

	flag.Parse()

	sets, err := generate()
	if err != nil {
		fmt.Fprintln(os.Stderr, "i18n:", err)

		return 1
	}

	if *check {
		return report(each(sets, codegen.Drift))
	}

	changed, err := each(sets, codegen.Write)
	for _, line := range changed {
		if _, printErr := fmt.Fprintln(os.Stdout, line); printErr != nil {
			return 1
		}
	}

	return report(nil, err)
}

func generate() ([]codegen.Files, error) {
	c, err := dictionaries.Build(locale.All())
	if err == nil {
		err = checkSources(c)
	}

	if err != nil {
		return nil, err
	}

	files, err := typescript.Render(c)
	if err != nil {
		return nil, err
	}

	numbers, err := typescript.Numbers(c)

	return []codegen.Files{files, numbers}, err
}

func each(sets []codegen.Files, step func(string, codegen.Files) ([]string, error)) ([]string, error) {
	var lines []string

	for _, files := range sets {
		found, err := step(".", files)

		lines = append(lines, found...)
		if err != nil {
			return lines, err
		}
	}

	return lines, nil
}

func checkSources(c dictionaries.Catalog) error {
	templateKeys, templateErr := keys.Check(views.FS, c.Args)
	goKeys, goErr := gocode.Check(".", reflect.TypeFor[api.Key](), c.Args)

	literals, scriptErr := frontend.Find(os.DirFS("."), "assets")
	if err := errors.Join(templateErr, goErr, scriptErr); err != nil {
		return err
	}

	return unused.Check(slices.Sorted(maps.Keys(c.Args)), slices.Concat(templateKeys, goKeys, []string{locale.NameKey}), literals)
}

func report(found []string, err error) int {
	if err != nil {
		fmt.Fprintln(os.Stderr, "i18n:", err)

		return 1
	}

	for _, issue := range found {
		fmt.Fprintln(os.Stderr, "i18n:", issue)
	}

	if len(found) > 0 {
		fmt.Fprintln(os.Stderr, "i18n: run go run ./tools/i18n")

		return 1
	}

	return 0
}
