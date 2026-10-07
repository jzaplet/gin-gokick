package templates

import (
	"slices"

	"gokick/app/core/locale"
	"gokick/app/core/reporting"
	"gokick/app/core/view"
	"gokick/app/core/vite"
	"gokick/app/internal/shared/config"
	"gokick/app/internal/shared/localeimages"
	"gokick/views"
)

func Parse(cfg *config.Config, assets *vite.Assets, browser *reporting.Browser, locales *locale.Set) (*view.Renderer, error) {
	byLocale, err := translations(locales.Locales())
	if err != nil {
		return nil, err
	}

	renderer, err := view.Parse(views.FS, funcs(cfg, assets, browser, locales, byLocale), byLocale)
	if err != nil {
		return nil, err
	}

	if err := assets.Check(renderer.Literals("vite")...); err != nil {
		return nil, err
	}

	files := slices.Concat(renderer.Literals("asset"), localeimages.Flags(locales.Languages()), localeimages.OGImages(locales.Locales()), dictionaryModules(locales.Locales()))
	if err := assets.CheckFiles(files...); err != nil {
		return nil, err
	}

	return renderer, nil
}
