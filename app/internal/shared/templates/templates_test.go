package templates_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"gokick/app/core/locale"
	"gokick/app/core/testkit"
	"gokick/app/core/vite"
	"gokick/app/core/vite/manifest"
	"gokick/app/internal/shared/config"
	"gokick/app/internal/shared/routertest"
	"gokick/app/internal/shared/templates"
	dictionaries "gokick/locale"
	"gokick/views"
)

func TestParseRefusesFilesMissingFromTheManifest(t *testing.T) {
	locales := routertest.Locales(t, "cs_CZ", "cs_CZ")
	missing := first(t, testkit.TemplateLiterals(t, views.FS, "asset"))

	err := parse(t, locales, without(routertest.Files(t, locales), missing))
	if errors.Is(err, manifest.ErrUnknownFile) == false || strings.HasSuffix(err.Error(), ": "+missing) == false {
		t.Errorf("without %s: error %v", missing, err)
	}
}

func TestParseRefusesALocaleWithoutOneOfItsFiles(t *testing.T) {
	locales := routertest.Locales(t, "cs_CZ", "cs_CZ", "en_US")
	for _, missing := range []string{"assets/img/flags/us.svg", "assets/img/og/en_US.png", "assets/shared/I18n/Dictionary/Locales/en_US.ts"} {
		err := parse(t, locales, without(routertest.Files(t, locales), missing))
		if errors.Is(err, manifest.ErrUnknownFile) == false || strings.HasSuffix(err.Error(), ": "+missing) == false {
			t.Errorf("without %s: error %v", missing, err)
		}
	}
}

func TestParseRefusesALocaleWithoutItsTexts(t *testing.T) {
	err := parse(t, routertest.Locales(t, "cs_CZ", "cs_CZ", "en_GB"), []string{"assets/app.css"})

	if errors.Is(err, dictionaries.ErrNoDictionary) == false || strings.HasSuffix(err.Error(), ": en_GB") == false {
		t.Errorf("error %v", err)
	}
}

func TestParseRefusesEntriesMissingFromTheManifest(t *testing.T) {
	locales := routertest.Locales(t, "cs_CZ", "cs_CZ")
	missing := first(t, testkit.TemplateLiterals(t, views.FS, "vite"))

	err := parse(t, locales, without(routertest.Files(t, locales), missing))
	if errors.Is(err, manifest.ErrUnknownEntry) == false || strings.HasSuffix(err.Error(), ": "+missing) == false {
		t.Errorf("without %s: error %v", missing, err)
	}
}

func parse(t *testing.T, locales *locale.Set, files []string) error {
	t.Helper()

	assets, err := vite.Load(testkit.Public(t, files...), testkit.DiscardLogs())
	if err != nil {
		t.Fatal(err)
	}

	_, err = templates.Parse(&config.Config{}, assets, nil, locales)

	return err
}

func first(t *testing.T, names []string) string {
	t.Helper()

	if len(names) == 0 {
		t.Fatal("the templates name nothing")
	}

	return names[0]
}

func without(names []string, missing string) []string {
	return slices.DeleteFunc(slices.Clone(names), func(name string) bool { return name == missing })
}
