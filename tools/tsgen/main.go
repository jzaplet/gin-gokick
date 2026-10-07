package main

import (
	"flag"
	"fmt"
	"os"

	"gokick/tools/tsgen/generator"
)

func main() {
	os.Exit(run())
}

func run() int {
	check := flag.Bool("check", false, "write nothing, fail when a generated file differs from the Go types")

	flag.Parse()

	if *check {
		return report(generator.Drift("."))
	}

	changed, err := generator.Generate(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	for _, line := range changed {
		if _, err := fmt.Fprintln(os.Stdout, line); err != nil {
			return 1
		}
	}

	return 0
}

func report(found []string, err error) int {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		return 1
	}

	for _, issue := range found {
		fmt.Fprintln(os.Stderr, "tsgen:", issue)
	}

	if len(found) > 0 {
		fmt.Fprintln(os.Stderr, "tsgen: run go run ./tools/tsgen")

		return 1
	}

	return 0
}
