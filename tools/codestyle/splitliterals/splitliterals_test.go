package splitliterals

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gokick/tools/codestyle/noalign"
)

var long = strings.Repeat("a", 60)

func TestFormatSplitsALongKeyedLiteral(t *testing.T) {
	src := "var v = T{Name: \"" + long + "\", Value: \"" + long + "\"}\n"

	want := "var v = T{\n\tName: \"" + long + "\",\n\n\tValue: \"" + long + "\",\n}\n"
	expectFormat(t, src, want)
}

func TestFormatSplitsTheOuterLiteralFirstAndTheInnerOnlyWhenStillTooLong(t *testing.T) {
	src := "var v = T{Name: \"x\", Inner: U{A: \"" + long + "\", B: \"" + long + "\"}}\n"

	want := "var v = T{\n\tName: \"x\",\n\n\tInner: U{\n\t\tA: \"" + long + "\",\n\n\t\tB: \"" + long + "\",\n\t},\n}\n"
	expectFormat(t, src, want)
}

func TestFormatSplitsALongMapLiteralInAStatement(t *testing.T) {
	src := "func f() {\n\tfor k, v := range map[string]string{\"a\": \"" + long + "\", \"b\": \"" + long + "\"} {\n\t\t_, _ = k, v\n\t}\n}\n"

	want := "func f() {\n\tfor k, v := range map[string]string{\n\t\t\"a\": \"" + long + "\",\n\n\t\t\"b\": \"" + long + "\",\n\t} {\n\t\t_, _ = k, v\n\t}\n}\n"
	expectFormat(t, src, want)
}

func TestFormatLeavesOtherCodeAlone(t *testing.T) {
	t.Run("short keyed literal", func(t *testing.T) { expectUnchanged(t, "var v = T{Name: \"x\", Value: \"y\"}\n") })
	t.Run("one field", func(t *testing.T) { expectUnchanged(t, "var v = T{Name: \""+long+long+"\"}\n") })
	t.Run("literal without keys", func(t *testing.T) { expectUnchanged(t, "var v = []string{\""+long+"\", \""+long+"\"}\n") })
	t.Run("comment inside", func(t *testing.T) { expectUnchanged(t, "var v = T{Name: \""+long+"\", /* c */ Value: \""+long+"\"}\n") })
	t.Run("already split", func(t *testing.T) {
		expectUnchanged(t, "var v = T{\n\tName: \""+long+"\",\n\n\tValue: \""+long+"\",\n}\n")
	})
}

func TestFormatMeasuresTabsAndRunes(t *testing.T) {
	fits := strings.Repeat("é", 44)
	expectUnchanged(t, "func f() {\n\t_ = T{Name: \""+fits+"\", Value: \""+fits+"\"}\n}\n")

	src := "func f() {\n\tif true {\n\t\t_ = T{Name: \"" + fits + "\", Value: \"" + fits + "\"}\n\t}\n}\n"
	if got := formatted(t, src); got == "package p\n\n"+src {
		t.Errorf("a line of 122 columns with two tabs stayed:\n%s", got)
	}
}

func TestFormattedCodeHasNoColumns(t *testing.T) {
	got := formatted(t, "var v = T{Name: \"x\", LongerName: \""+long+"\", Value: map[string]int{\"a\": 1, \"bbbbbb\": 2}, Other: \""+long+"\"}\n")

	if issues := noalign.Find("case.go", []byte(got)); len(issues) != 0 {
		t.Errorf("columns %v in\n%s", issues, got)
	}

	if again := formatted(t, strings.TrimPrefix(got, "package p\n\n")); again != got {
		t.Errorf("a second run changed\n%s\ninto\n%s", got, again)
	}
}

func TestRewriteTouchesOnlyFilesWithLongLiterals(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "long.go"), "package p\n\nvar v = T{Name: \""+long+"\", Value: \""+long+"\"}\n")
	write(t, filepath.Join(root, "short.go"), "package p\n\nvar v = T{Name: \"x\", Value: \"y\"}\n")

	if issues, err := Issues(root); err != nil || len(issues) != 1 || issues[0].Pos.Filename != filepath.Join(root, "long.go") || issues[0].Pos.Line != 3 {
		t.Errorf("issues %v, err %v before formatting", issues, err)
	}

	changed, err := Rewrite(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(changed) != 1 || changed[0] != filepath.Join(root, "long.go") {
		t.Errorf("changed %v", changed)
	}

	if issues, err := Issues(root); err != nil || len(issues) != 0 {
		t.Errorf("issues %v, err %v after formatting", issues, err)
	}
}

func expectFormat(t *testing.T, src, want string) {
	t.Helper()

	if got := formatted(t, src); got != "package p\n\n"+want {
		t.Errorf("got\n%s\nwant\npackage p\n\n%s", got, want)
	}
}

func expectUnchanged(t *testing.T, src string) {
	t.Helper()
	expectFormat(t, src, src)
}

func formatted(t *testing.T, src string) string {
	t.Helper()

	out, err := splitLong([]byte("package p\n\n" + src))
	if err != nil {
		t.Fatal(err)
	}

	return string(out)
}

func write(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
