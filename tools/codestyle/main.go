package main

import (
	"flag"
	"fmt"
	"os"
	"slices"

	"gokick/tools/codestyle/gofiles"
	"gokick/tools/codestyle/htmllayout"
	"gokick/tools/codestyle/noalign"
	"gokick/tools/codestyle/nonegation"
	"gokick/tools/codestyle/splitliterals"
)

const columnsHint = "keep a keyed literal on one line, split struct fields and consts by a blank line or give them one shared type"

func main() {
	os.Exit(run())
}

func run() int {
	check := flag.Bool("check", false, "write nothing, fail on a column, on a keyed literal wider than the limit, on a negation with ! or on a template not laid out")

	flag.Parse()

	if *check == false {
		if err := rewrite(); err != nil {
			return fail(err)
		}
	}

	columns, err := noalign.Issues(".")
	if err != nil {
		return fail(err)
	}

	long, err := splitliterals.Issues(".")
	if err != nil {
		return fail(err)
	}

	negations, err := nonegation.Issues(".")
	if err != nil {
		return fail(err)
	}

	templates, err := htmllayout.Issues(".")
	if err != nil {
		return fail(err)
	}

	return max(report(columns, columnsHint), report(slices.Concat(long, negations, templates), "run go run ./tools/codestyle"))
}

func rewrite() error {
	compared, err := nonegation.Rewrite(".")
	if err != nil {
		return err
	}

	split, err := splitliterals.Rewrite(".")
	if err != nil {
		return err
	}

	laidOut, err := htmllayout.Rewrite(".")
	for _, path := range slices.Concat(compared, split, laidOut) {
		if _, printErr := fmt.Fprintln(os.Stdout, path); printErr != nil {
			return printErr
		}
	}

	return err
}

func report(issues []gofiles.Issue, hint string) int {
	for _, issue := range issues {
		fmt.Fprintln(os.Stderr, "codestyle:", issue)
	}

	if len(issues) == 0 {
		return 0
	}

	fmt.Fprintln(os.Stderr, "codestyle:", hint)

	return 1
}

func fail(err error) int {
	fmt.Fprintln(os.Stderr, "codestyle:", err)

	return 1
}
