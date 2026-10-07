package vite

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"gokick/app/core/vite/devserver"
	"gokick/app/core/vite/manifest"
	"gokick/app/core/vite/tags"
)

const buildDir = "build"

const manifestFile = buildDir + "/manifest.json"

const base = "/" + buildDir + "/"

const HashedAssets = buildDir + "/assets/"

const hotFile = "public/.hot"

type Assets struct {
	source source

	dev *devserver.Server
}

type source interface {
	Entry(name string) (tags.Set, error)
	Module(name string) (tags.Set, error)
	URL(name string) (string, error)
}

func Load(public fs.FS, logger *slog.Logger) (*Assets, error) {
	return load(public, hotFile, logger)
}

func load(public fs.FS, hot string, logger *slog.Logger) (*Assets, error) {
	target, err := readHotFile(hot)
	if err != nil {
		return nil, err
	}

	if target != nil {
		dev := devserver.New(target, base, logger)

		return &Assets{source: dev, dev: dev}, nil
	}

	build, err := manifest.Read(public, manifestFile, base)
	if err != nil {
		return nil, err
	}

	return &Assets{source: build}, nil
}

func (a *Assets) Register(router gin.IRoutes) {
	if a.dev != nil {
		a.dev.Register(router)
	}
}

func (a *Assets) Check(entries ...string) error {
	return check(manifest.ErrUnknownEntry, entries, func(name string) error {
		_, err := a.source.Entry(name)

		return err
	})
}

func (a *Assets) CheckFiles(names ...string) error {
	return check(manifest.ErrUnknownFile, names, func(name string) error {
		_, err := a.source.URL(name)

		return err
	})
}

func (a *Assets) URL(name string) (string, error) {
	return a.source.URL(name)
}

func (a *Assets) Tags(nonce string, entries ...string) (template.HTML, error) {
	var set tags.Set
	if a.dev != nil {
		set.Add(a.dev.Client())
	}

	for _, entry := range entries {
		entryTags, err := a.source.Entry(entry)
		if err != nil {
			return "", err
		}

		set.Merge(entryTags)
	}

	return set.HTML(nonce), nil
}

func (a *Assets) Preload(nonce string, modules ...string) (template.HTML, error) {
	var set tags.Set

	for _, module := range modules {
		moduleTags, err := a.source.Module(module)
		if err != nil {
			return "", err
		}

		set.Merge(moduleTags)
	}

	return set.HTML(nonce), nil
}

func readHotFile(name string) (*url.URL, error) {
	hot, err := os.ReadFile(name)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("vite: %w", err)
	}

	target, err := url.Parse(strings.TrimSpace(string(hot)))
	if err != nil || target.Host == "" {
		return nil, fmt.Errorf("vite: %s holds no dev server URL: %q", name, hot)
	}

	return target, nil
}

func check(unknown error, names []string, find func(name string) error) error {
	var missing []string
	for _, name := range names {
		if err := find(name); err == nil || slices.Contains(missing, name) {
			continue
		}

		missing = append(missing, name)
	}

	if len(missing) == 0 {
		return nil
	}

	slices.Sort(missing)

	return fmt.Errorf("%w: %s", unknown, strings.Join(missing, ", "))
}
