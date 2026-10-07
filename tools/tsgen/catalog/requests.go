package catalog

import (
	"fmt"

	"gokick/tools/tsgen/gosource"
)

const general = "general"

const Index = "[]"

func (c *Catalog) addErrorPaths() error {
	dtos := make(map[gosource.TypeRef]*DTO, len(c.DTOs))
	for i := range c.DTOs {
		dtos[c.DTOs[i].Ref] = &c.DTOs[i]
	}

	for i := range c.DTOs {
		d := &c.DTOs[i]
		if d.Dir.Request == false {
			continue
		}

		paths, err := errorPaths(dtos, d, "", map[gosource.TypeRef]bool{})
		if err != nil {
			return fmt.Errorf("request %s: %w", d.Ref, err)
		}

		d.ErrorPaths = append([]string{general}, paths...)
	}

	return nil
}

func errorPaths(dtos map[gosource.TypeRef]*DTO, d *DTO, prefix string, open map[gosource.TypeRef]bool) ([]string, error) {
	if open[d.Ref] {
		return nil, fmt.Errorf("%s contains itself, so its error paths never end", d.Ref)
	}

	open[d.Ref] = true
	defer delete(open, d.Ref)

	var paths []string

	for _, f := range d.Fields {
		path := prefix + f.Name
		if path == general {
			return nil, fmt.Errorf("the json name %q is kept for the error outside the fields", general)
		}

		paths = append(paths, path)

		nested, ok := dtos[f.Shape.Ref]
		if f.Shape.Array {
			if ok && f.Dive == false {
				return nil, fmt.Errorf("%s: a list of %s needs dive in its binding tag, or its items go unchecked", path, f.Shape.Ref)
			}

			path += Index
			paths = append(paths, path)
		}

		if ok == false {
			continue
		}

		inner, err := errorPaths(dtos, nested, path+".", open)
		if err != nil {
			return nil, err
		}

		paths = append(paths, inner...)
	}

	return paths, nil
}
