package codegen

import (
	"bytes"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Files struct {
	Header string

	Dir string

	Content map[string][]byte
}

func Write(root string, files Files) ([]string, error) {
	fsys := os.DirFS(root)
	var changed []string

	for _, rel := range slices.Sorted(maps.Keys(files.Content)) {
		current, err := fs.ReadFile(fsys, rel)
		if err == nil && bytes.Equal(current, files.Content[rel]) {
			continue
		}

		target := filepath.Join(root, filepath.FromSlash(rel))

		err = os.MkdirAll(filepath.Dir(target), 0o750)
		if err == nil {
			err = os.WriteFile(target, files.Content[rel], 0o600)
		}

		if err != nil {
			return changed, err
		}

		changed = append(changed, "wrote "+rel)
	}

	orphans, err := orphans(root, files)
	if err != nil {
		return changed, err
	}

	for _, rel := range orphans {
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			return changed, err
		}

		changed = append(changed, "removed "+rel)
	}

	return changed, nil
}

func Drift(root string, files Files) ([]string, error) {
	fsys := os.DirFS(root)
	var found []string

	for _, rel := range slices.Sorted(maps.Keys(files.Content)) {
		current, err := fs.ReadFile(fsys, rel)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			found = append(found, rel+" is missing")
		case err != nil:
			return nil, err
		case bytes.Equal(current, files.Content[rel]) == false:
			found = append(found, rel+" is out of date")
		}
	}

	orphans, err := orphans(root, files)
	for _, rel := range orphans {
		found = append(found, rel+" is no longer generated")
	}

	return found, err
}

func orphans(root string, files Files) ([]string, error) {
	fsys := os.DirFS(root)
	var found []string

	err := fs.WalkDir(fsys, files.Dir, func(rel string, entry fs.DirEntry, err error) error {
		if _, live := files.Content[rel]; err != nil || live || entry.IsDir() || strings.HasSuffix(rel, ".ts") == false {
			return err
		}

		content, err := fs.ReadFile(fsys, rel)
		if err == nil && bytes.HasPrefix(content, []byte(files.Header)) {
			found = append(found, rel)
		}

		return err
	})
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	return found, err
}
