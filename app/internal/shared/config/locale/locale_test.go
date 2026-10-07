package locale

import (
	"reflect"
	"testing"
)

func TestLocalesDefaultToTheDefaultLocale(t *testing.T) {
	t.Setenv("DEFAULT_LOCALE", "en_US")
	t.Setenv("LOCALES", "")

	cfg, err := Read()
	if want := (Config{
		Default: "en_US",

		Locales: []string{"en_US"},
	}); err != nil || reflect.DeepEqual(cfg, want) == false {
		t.Errorf("got %+v, %v", cfg, err)
	}
}

func TestReadsTheLocalesInTheirOrder(t *testing.T) {
	t.Setenv("DEFAULT_LOCALE", "cs_CZ")
	t.Setenv("LOCALES", "en_GB,cs_CZ,en_US")

	cfg, err := Read()
	if want := (Config{
		Default: "cs_CZ",

		Locales: []string{"en_GB", "cs_CZ", "en_US"},
	}); err != nil || reflect.DeepEqual(cfg, want) == false {
		t.Errorf("got %+v, %v", cfg, err)
	}
}
