package gocode

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"gokick/app/core/i18n/icu"
	"gokick/tools/i18n/gocode/testdata/fixture/api"
)

var texts = map[string]map[string]icu.Kind{
	"request.internal": {},

	"user.taken": {},

	"user.gone": {},

	"user.local": {},
}

func TestCheckRefusesKeysFromNoPackageConstantAndConstantsWithoutText(t *testing.T) {
	keys, err := Check("testdata/fixture", reflect.TypeFor[api.Key](), texts)
	if err == nil {
		t.Fatal("Check = nil")
	}

	if want := []string{"request.internal", "user.taken", "user.gone", "user.untranslated", "user.local"}; slices.Equal(keys, want) == false {
		t.Errorf("keys = %v, want %v", keys, want)
	}

	want := []string{
		"handler/handler.go:10:2: " + ErrNoText.Error() + `: KeyUntranslated "user.untranslated"`,
		"handler/handler.go:16:8: " + ErrLocalConstant.Error() + `: keyLocal`,
		"handler/handler.go:24:12: " + ErrNotConstant.Error() + `: "user.literal"`,
		"handler/handler.go:25:12: " + ErrNotConstant.Error() + `: api.Key(name)`,
		"handler/handler.go:26:12: " + ErrNotConstant.Error() + `: api.Key("user.converted")`,
		"handler/handler.go:27:12: " + ErrNotConstant.Error() + `: KeyTaken + ".more"`,
		"handler/handler.go:28:12: " + ErrNotConstant.Error() + `: prefix + "x"`,
		"handler/handler.go:29:9: " + ErrNotConstant.Error() + `: "user.field"`,
		"handler/handler.go:31:12: " + ErrNotConstant.Error() + `: "user.compared"`,
		"handler/handler.go:43:9: " + ErrNotConstant.Error() + `: "user.returned"`,
	}
	if got := strings.Split(err.Error(), "\n"); slices.Equal(got, want) == false {
		t.Errorf("Check =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestCheckRefusesAKeyTypeWithoutConstants(t *testing.T) {
	_, err := Check("testdata/fixture", reflect.TypeFor[api.Message](), texts)
	if errors.Is(err, ErrNoKeys) == false {
		t.Errorf("Check = %v, want %v", err, ErrNoKeys)
	}
}

func TestCheckFailsWhenGoListFails(t *testing.T) {
	if _, err := Check("testdata/missing", reflect.TypeFor[api.Key](), texts); err == nil || strings.Contains(err.Error(), "go list") == false {
		t.Errorf("Check = %v", err)
	}
}
