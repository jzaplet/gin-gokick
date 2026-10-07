package clientinfo_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"gokick/app/internal/auth/session/clientinfo"
)

func TestUserAgentIsValidTextWithoutNULOfAtMostTheLimit(t *testing.T) {
	got := clientinfo.UserAgent("Mozilla\x00 \xff" + strings.Repeat("é", clientinfo.MaxUserAgent))

	if utf8.ValidString(got) == false || strings.ContainsRune(got, 0) || utf8.RuneCountInString(got) != clientinfo.MaxUserAgent {
		t.Errorf("got %q", got)
	}

	if strings.HasPrefix(got, "Mozilla é") == false {
		t.Errorf("got %q, want the text without the invalid bytes", got[:12])
	}
}
