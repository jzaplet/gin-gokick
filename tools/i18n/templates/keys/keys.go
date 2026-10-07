package keys

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"slices"
	"text/template/parse"

	"gokick/app/core/i18n/icu"
	"gokick/app/core/view/literals"
	"gokick/tools/i18n/templates/baretext"
)

var ErrTemplate = errors.New("a template does not parse")

var ErrKey = errors.New("t takes no literal key")

var ErrUnknownKey = errors.New("t names a key the dictionaries lack")

var ErrParams = errors.New("t passes other params than its message takes")

func Check(fsys fs.FS, args map[string]map[string]icu.Kind) ([]string, error) {
	var keys []string
	var issues []error

	err := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || path.Ext(name) != ".html" {
			return err
		}

		content, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}

		fileKeys, fileIssues := checkFile(name, string(content), args)
		keys = append(keys, fileKeys...)
		issues = append(issues, fileIssues...)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return keys, errors.Join(issues...)
}

func checkFile(name, content string, args map[string]map[string]icu.Kind) ([]string, []error) {
	file := parse.New(name)
	file.Mode = parse.SkipFuncCheck

	trees := map[string]*parse.Tree{}
	if _, err := file.Parse(content, "", "", trees); err != nil {
		return nil, []error{fmt.Errorf("%w: %w", ErrTemplate, err)}
	}

	var keys []string
	var issues []error

	for _, define := range slices.Sorted(maps.Keys(trees)) {
		tree := trees[define]

		issues = append(issues, baretext.Check(tree)...)
		for _, call := range literals.Calls(tree.Root, "t") {
			key, err := checkCall(call, args)
			if err != nil {
				location, _ := tree.ErrorContext(call)
				issues = append(issues, fmt.Errorf("%s: %w", location, err))
			}

			if key != "" {
				keys = append(keys, key)
			}
		}
	}

	return keys, issues
}

func checkCall(call *parse.CommandNode, args map[string]map[string]icu.Kind) (string, error) {
	if len(call.Args) < 2 {
		return "", fmt.Errorf("%w: %s", ErrKey, call)
	}

	key, ok := call.Args[1].(*parse.StringNode)
	if ok == false {
		return "", fmt.Errorf("%w: %s", ErrKey, call)
	}

	kinds, ok := args[key.Text]
	if ok == false {
		return key.Text, fmt.Errorf("%w: %s", ErrUnknownKey, key.Text)
	}

	names, ok := paramNames(call.Args[2:])
	if ok == false || slices.Equal(names, slices.Sorted(maps.Keys(kinds))) == false {
		return key.Text, fmt.Errorf("%w: %s", ErrParams, call)
	}

	return key.Text, nil
}

func paramNames(pairs []parse.Node) ([]string, bool) {
	if len(pairs)%2 != 0 {
		return nil, false
	}

	names := make([]string, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		name, ok := pairs[i].(*parse.StringNode)
		if ok == false || slices.Contains(names, name.Text) {
			return nil, false
		}

		names = append(names, name.Text)
	}

	slices.Sort(names)

	return names, true
}
