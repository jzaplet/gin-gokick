package reporting

import (
	"cmp"
	"html/template"
)

type Browser struct {
	dsn string

	environment string

	release string
}

func NewBrowser(dsn, environment, release string) *Browser {
	if dsn == "" {
		return nil
	}

	return &Browser{dsn: dsn, environment: environment, release: cmp.Or(release, buildRevision())}
}

func (b *Browser) Meta() template.HTML {
	if b == nil {
		return ""
	}

	attr := template.HTMLEscapeString

	return template.HTML(`<meta name="sentry" data-dsn="` + attr(b.dsn) + `" data-environment="` + attr(b.environment) + `" data-release="` + attr(b.release) + `">`)
}
