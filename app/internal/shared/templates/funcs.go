package templates

import (
	"html/template"
	"strconv"
	"time"

	"gokick/app/core/locale"
	"gokick/app/core/locale/hreflang"
	"gokick/app/core/locale/posix"
	"gokick/app/core/reporting"
	"gokick/app/core/view"
	"gokick/app/core/vite"
	"gokick/app/internal/shared/config"
	"gokick/app/internal/shared/config/tracking"
	"gokick/app/internal/shared/csp"
	"gokick/app/internal/shared/localeimages"
	dictionaries "gokick/locale"
)

func funcs(cfg *config.Config, assets *vite.Assets, browser *reporting.Browser, locales *locale.Set, byLocale map[string]view.Translate) template.FuncMap {
	return template.FuncMap{
		"vite": assets.Tags,

		"asset": assets.URL,

		"sentry": browser.Meta,

		"tracking": func() *tracking.Config { return &cfg.Tracking },

		"umamiScript": func() string { return csp.UmamiScript },

		"year": func() string { return strconv.Itoa(time.Now().Year()) },

		"flag": func(l posix.Locale) (string, error) { return assets.URL(localeimages.Flag(l)) },

		"ogImage": func(l posix.Locale) (string, error) { return assets.URL(localeimages.OGImage(l)) },

		"home": func(l posix.Locale) string { return hreflang.Path(l.Language(), locales.DefaultLanguage(), "") },

		"languageName": languageName(byLocale),

		"dictionary": func(nonce string, l posix.Locale) (template.HTML, error) {
			return assets.Preload(nonce, dictionaries.Module(l.String()))
		},
	}
}
