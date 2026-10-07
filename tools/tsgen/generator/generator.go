package generator

import (
	"fmt"
	"path/filepath"

	"gokick/tools/codegen"
	"gokick/tools/tsgen/catalog"
	"gokick/tools/tsgen/gosource"
	"gokick/tools/tsgen/typescript"
)

func Build(dir, basePkg string) (codegen.Files, error) {
	sources, err := gosource.Parse(dir, basePkg)
	if err != nil {
		return codegen.Files{}, err
	}

	c, err := catalog.Collect(sources)
	if err != nil {
		return codegen.Files{}, err
	}

	content, err := typescript.Render(c)

	return codegen.Files{Header: typescript.Header, Dir: "assets", Content: content}, err
}

func Generate(root string) ([]string, error) {
	files, err := moduleFiles(root)
	if err != nil {
		return nil, err
	}

	return codegen.Write(root, files)
}

func Drift(root string) ([]string, error) {
	files, err := moduleFiles(root)
	if err != nil {
		return nil, err
	}

	return codegen.Drift(root, files)
}

func moduleFiles(root string) (codegen.Files, error) {
	files, err := Build(filepath.Join(root, "app"), "gokick/app")
	if err != nil {
		return codegen.Files{}, fmt.Errorf("tsgen: %w", err)
	}

	return files, nil
}
