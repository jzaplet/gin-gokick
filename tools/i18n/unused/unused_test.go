package unused

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"gokick/tools/i18n/frontend"
)

var keys = []string{"home.title", "login.title", "request.not_found", "tracking.ga4", "tracking.meta_pixel"}

func TestCheckAcceptsKeysEverySourceUses(t *testing.T) {
	err := Check(keys, []string{"home.title", "request.not_found"}, []frontend.Literal{
		{Pos: "assets/a.ts:1:1", Text: "login.title"},

		{Pos: "assets/b.ts:1:1", Text: "tracking.ga4"},

		{Pos: "assets/b.ts:2:1", Text: "tracking.meta_pixel"},

		{Pos: "assets/c.ts:1:1", Text: "dictionary.", Built: true},

		{Pos: "assets/c.ts:2:1", Text: "."},
	})
	if err != nil {
		t.Errorf("Check = %v", err)
	}
}

func TestCheckRefusesUnusedAndBuiltKeys(t *testing.T) {
	err := Check(keys, []string{"home.title", "request.not_found"}, []frontend.Literal{
		{Pos: "assets/a.ts:1:1", Text: "login.title"},

		{Pos: "assets/b.ts:4:9", Text: "tracking.", Built: true},

		{Pos: "assets/b.ts:5:9", Text: "tracking."},

		{Pos: "assets/b.ts:6:9", Text: "tracking.ga4", Built: true},
	})

	want := []string{
		`assets/b.ts:4:9: ` + ErrBuilt.Error() + `: "tracking."`,
		`assets/b.ts:5:9: ` + ErrBuilt.Error() + `: "tracking."`,
		ErrUnused.Error() + ": tracking.ga4",
		ErrUnused.Error() + ": tracking.meta_pixel",
	}
	if errors.Is(err, ErrUnused) == false || errors.Is(err, ErrBuilt) == false || slices.Equal(strings.Split(err.Error(), "\n"), want) == false {
		t.Errorf("Check =\n%v\nwant\n%s", err, strings.Join(want, "\n"))
	}
}
