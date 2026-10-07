package locale

import (
	"errors"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestEveryDictionaryFileRegistersItsLocale(t *testing.T) {
	files, err := filepath.Glob("[a-z][a-z]_[A-Z][A-Z].go")
	if err != nil || len(files) == 0 {
		t.Fatalf("Glob = %v, %v", files, err)
	}

	var names []string
	for _, file := range files {
		names = append(names, strings.TrimSuffix(file, ".go"))
	}

	if got := slices.Sorted(maps.Keys(All())); slices.Equal(got, names) == false {
		t.Errorf("All() = %v, files %v", got, names)
	}
}

func TestRequireAcceptsOnlyLocalesWithADictionary(t *testing.T) {
	for name := range All() {
		if err := Require([]string{name}); err != nil {
			t.Errorf("Require(%s) = %v", name, err)
		}
	}

	if err := Require([]string{"zz_ZZ"}); errors.Is(err, ErrNoDictionary) == false {
		t.Errorf("Require(zz_ZZ) = %v", err)
	}
}
