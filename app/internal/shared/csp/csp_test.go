package csp

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"gokick/app/internal/shared/config/tracking"
)

const googleTagHost = "https://www.googletagmanager.com"

const turnstileHost = "https://challenges.cloudflare.com"

var allTools = tracking.Config{
	UmamiWebsiteID: "0192f3a4-5b6c-7d8e-9f01-23456789abcd",

	GA4MeasurementID: "G-AB12CD34EF",

	GoogleAdsID: "AW-123456789",

	MetaPixelID: "1234567890123456",
}

func TestPolicyAllowsOnlyTheEnabledTools(t *testing.T) {
	loader := string(ScriptLoader)
	for name, tt := range map[string]struct {
		tools tracking.Config

		scripts []string

		frames []string

		trustedTypes []string
	}{
		"no tool": {
			tools: tracking.Config{},

			scripts: []string{turnstileHost},

			frames: []string{turnstileHost},

			trustedTypes: []string{"vue"},
		},

		"GA4": {
			tools: tracking.Config{GA4MeasurementID: "G-AB12CD34EF"},

			scripts: []string{googleTagHost, turnstileHost},

			frames: []string{googleTagHost, turnstileHost},

			trustedTypes: []string{"goog#html", "'allow-duplicates'", "vue", loader},
		},

		"Google Ads": {
			tools: tracking.Config{GoogleAdsID: "AW-123456789"},

			scripts: []string{googleTagHost, "https://www.googleadservices.com", "https://www.google.com", turnstileHost},

			frames: []string{googleTagHost, turnstileHost},

			trustedTypes: []string{"goog#html", "'allow-duplicates'", "vue", loader},
		},

		"Meta Pixel": {
			tools: tracking.Config{MetaPixelID: "1234567890123456"},

			scripts: []string{"https://connect.facebook.net", turnstileHost},

			frames: []string{turnstileHost},

			trustedTypes: []string{"connect.facebook.net/fbevents", "facebook.com/signals/iwl", "vue", loader},
		},

		"Umami": {
			tools: tracking.Config{UmamiWebsiteID: "0192f3a4-5b6c-7d8e-9f01-23456789abcd"},

			scripts: []string{umami, turnstileHost},

			frames: []string{turnstileHost},

			trustedTypes: []string{"vue"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			expectDirective(t, &tt.tools, "script-src", append([]string{"'nonce-abc'", "'strict-dynamic'"}, tt.scripts...))
			expectDirective(t, &tt.tools, "frame-src", tt.frames)
			expectDirective(t, &tt.tools, "trusted-types", tt.trustedTypes)
		})
	}
}

func TestScriptURLsComeFromTheScriptSources(t *testing.T) {
	sources := directive(t, &allTools, "script-src")

	scripts := scriptURLs(t)
	if len(scripts) == 0 {
		t.Fatal("the package has no ScriptURL constants")
	}

	for _, script := range scripts {
		parsed, err := url.Parse(script)
		if err != nil {
			t.Fatal(err)
		}

		plain := parsed.Scheme == "https" && parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == ""
		if plain == false || slices.Contains(sources, parsed.Scheme+"://"+parsed.Host) == false {
			t.Errorf("%s is no plain HTTPS address of script-src %v", script, sources)
		}
	}
}

func expectDirective(t *testing.T, tools *tracking.Config, name string, want []string) {
	t.Helper()

	if got := directive(t, tools, name); slices.Equal(slices.Sorted(slices.Values(got)), slices.Sorted(slices.Values(want))) == false {
		t.Errorf("%s %v, want %v", name, got, want)
	}
}

func directive(t *testing.T, tools *tracking.Config, name string) []string {
	t.Helper()

	for part := range strings.SplitSeq(Policy(tools).Header().Value("abc"), "; ") {
		if fields := strings.Fields(part); fields[0] == name {
			return fields[1:]
		}
	}

	t.Fatalf("CSP has no %s", name)

	return nil
}

func scriptURLs(t *testing.T) []string {
	t.Helper()

	names, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	var scripts []string

	for _, name := range names {
		file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}

		for node := range ast.Preorder(file) {
			if spec, ok := node.(*ast.ValueSpec); ok && isScriptURL(spec.Type) {
				scripts = append(scripts, literals(t, spec)...)
			}
		}
	}

	return scripts
}

func isScriptURL(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)

	return ok && ident.Name == "ScriptURL"
}

func literals(t *testing.T, spec *ast.ValueSpec) []string {
	t.Helper()

	values := make([]string, 0, len(spec.Values))
	for _, expr := range spec.Values {
		lit, ok := expr.(*ast.BasicLit)
		if ok == false || lit.Kind != token.STRING {
			t.Fatalf("%s is no string literal", spec.Names[0].Name)
		}

		value, err := strconv.Unquote(lit.Value)
		if err != nil {
			t.Fatal(err)
		}

		values = append(values, value)
	}

	return values
}
