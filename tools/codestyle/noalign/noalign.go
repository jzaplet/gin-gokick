package noalign

import (
	"go/scanner"
	"go/token"

	"gokick/tools/codestyle/gofiles"
)

func Issues(root string) ([]gofiles.Issue, error) {
	var issues []gofiles.Issue

	err := gofiles.Walk(root, func(path string, src []byte) error {
		issues = append(issues, Find(path, src)...)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return issues, nil
}

func Find(filename string, src []byte) []gofiles.Issue {
	fset := token.NewFileSet()
	file := fset.AddFile(filename, -1, len(src))
	var s scanner.Scanner
	s.Init(file, src, nil, scanner.ScanComments)

	var issues []gofiles.Issue
	prevEndLine, prevEndColumn := 0, 0

	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			return issues
		}

		if insertedSemicolon(tok, lit) {
			continue
		}

		start := fset.Position(pos)
		if start.Line == prevEndLine && start.Column > prevEndColumn+1 {
			issues = append(issues, gofiles.Issue{Pos: start, Msg: "aligned into a column by gofmt"})
		}

		text := lit
		if text == "" {
			text = tok.String()
		}

		end := fset.Position(pos + token.Pos(len(text)))
		prevEndLine, prevEndColumn = end.Line, end.Column
	}
}

func insertedSemicolon(tok token.Token, lit string) bool {
	return tok == token.SEMICOLON && lit == "\n"
}
