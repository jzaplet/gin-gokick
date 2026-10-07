package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"

	"gokick/app/core/vite/tags"
)

var ErrNoManifest = errors.New("vite: no manifest.json; build the assets, or start the dev server before the app")

var ErrUnknownEntry = errors.New("vite: entry is not in the manifest")

var ErrUnknownFile = errors.New("vite: file is not in the manifest")

type Manifest struct {
	entries map[string]tags.Set

	files map[string]string

	chunks map[string]chunk

	base string
}

type chunk struct {
	File string `json:"file"`

	CSS []string `json:"css"`

	Imports []string `json:"imports"`

	IsEntry bool `json:"isEntry"`
}

func Read(public fs.FS, name, base string) (*Manifest, error) {
	data, err := fs.ReadFile(public, name)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNoManifest
	}

	if err != nil {
		return nil, fmt.Errorf("vite: %w", err)
	}

	var chunks map[string]chunk
	if err := json.Unmarshal(data, &chunks); err != nil {
		return nil, fmt.Errorf("vite: %s: %w", name, err)
	}

	m := &Manifest{entries: map[string]tags.Set{}, files: map[string]string{}, chunks: chunks, base: base}
	for source, c := range chunks {
		m.files[source] = base + c.File
		if c.IsEntry {
			var set tags.Set
			collect(&set, chunks, c, base)
			set.Add(base + c.File)
			m.entries[source] = set
		}
	}

	return m, nil
}

func (m *Manifest) Entry(name string) (tags.Set, error) {
	set, ok := m.entries[name]
	if ok == false {
		return set, fmt.Errorf("%w: %s", ErrUnknownEntry, name)
	}

	return set, nil
}

func (m *Manifest) Module(name string) (tags.Set, error) {
	var set tags.Set

	c, ok := m.chunks[name]
	if ok == false {
		return set, fmt.Errorf("%w: %s", ErrUnknownFile, name)
	}

	collect(&set, m.chunks, c, m.base)
	set.Preload(m.base + c.File)

	return set, nil
}

func (m *Manifest) URL(name string) (string, error) {
	file, ok := m.files[name]
	if ok == false {
		return "", fmt.Errorf("%w: %s", ErrUnknownFile, name)
	}

	return file, nil
}

func collect(set *tags.Set, chunks map[string]chunk, c chunk, base string) {
	for _, name := range c.Imports {
		imported := chunks[name]
		if set.Preload(base + imported.File) {
			collect(set, chunks, imported, base)
		}
	}

	for _, file := range c.CSS {
		set.Style(base + file)
	}
}
