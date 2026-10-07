package nonegation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRewriteComparesWithFalse(t *testing.T) {
	t.Run("a variable", func(t *testing.T) { expectRewrite(t, "_ = !ok", "_ = ok == false") })
	t.Run("a call", func(t *testing.T) {
		expectRewrite(t, "_ = !strings.HasPrefix(s, p)", "_ = strings.HasPrefix(s, p) == false")
	})
	t.Run("a field", func(t *testing.T) { expectRewrite(t, "_ = !c.ok && d", "_ = c.ok == false && d") })
	t.Run("a group", func(t *testing.T) { expectRewrite(t, "_ = !(a && b)", "_ = (a && b) == false") })
	t.Run("a condition", func(t *testing.T) { expectRewrite(t, "if !ok {\n\t\treturn\n\t}", "if ok == false {\n\t\treturn\n\t}") })
}

func TestRewriteInvertsAComparison(t *testing.T) {
	t.Run("equal", func(t *testing.T) { expectRewrite(t, "_ = !(a == b)", "_ = a != b") })
	t.Run("not equal", func(t *testing.T) { expectRewrite(t, "_ = !(a != b)", "_ = a == b") })
	t.Run("less", func(t *testing.T) { expectRewrite(t, "_ = !(a < b)", "_ = (a < b) == false") })
}

func TestRewriteInvertsTheComparisonAroundANegation(t *testing.T) {
	t.Run("equal", func(t *testing.T) { expectRewrite(t, "_ = a == !b", "_ = a != b") })
	t.Run("not equal first", func(t *testing.T) { expectRewrite(t, "_ = !a != b", "_ = a == b") })
	t.Run("a group", func(t *testing.T) { expectRewrite(t, "_ = a == !(b && c)", "_ = a != (b && c)") })
}

func TestRewriteKeepsTheMeaningInsideAnArithmeticComparison(t *testing.T) {
	expectRewrite(t, "_ = a < !b", "_ = a < (b == false)")
}

func TestRewriteStartsWithTheInnermostNegation(t *testing.T) {
	expectRewrite(t, "_ = !(a && !b)", "_ = (a && b == false) == false")
}

func TestRewriteTouchesOnlyFilesWithNegations(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "negated.go"), "package p\n\nvar v = !w\n")
	write(t, filepath.Join(root, "compared.go"), "package p\n\nvar v = w == false\n")

	changed, err := Rewrite(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(changed) != 1 || changed[0] != filepath.Join(root, "negated.go") {
		t.Errorf("changed %v", changed)
	}

	if issues, err := Issues(root); err != nil || len(issues) != 0 {
		t.Errorf("issues %v, err %v after the rewrite", issues, err)
	}
}

func TestIssuesPointAtEveryNegation(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "negated.go"), "package p\n\nvar v = !w && !x\n")

	issues, err := Issues(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(issues) != 2 || issues[0].Pos.Column != 9 || issues[1].Pos.Column != 15 {
		t.Errorf("issues %v", issues)
	}
}

func expectRewrite(t *testing.T, statement, want string) {
	t.Helper()

	src := "package p\n\nfunc f() {\n\t" + statement + "\n}\n"

	got, err := compareExplicitly([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "package p\n\nfunc f() {\n\t"+want+"\n}\n" {
		t.Errorf("got\n%s\nwant the statement\n%s", got, want)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
