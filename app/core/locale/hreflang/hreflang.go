package hreflang

import (
	"net/http"

	"gokick/app/core/pageurl"
)

const defaultLink = "x-default"

type Link struct {
	Lang string

	URL string
}

func Links(req *http.Request, languages []string, defaultLanguage, page string) []Link {
	if len(languages) < 2 {
		return nil
	}

	origin := pageurl.Origin(req)

	links := make([]Link, 0, len(languages)+1)
	for _, lang := range languages {
		links = append(links, Link{Lang: lang, URL: origin + Path(lang, defaultLanguage, page)})
	}

	return append(links, Link{Lang: defaultLink, URL: origin + Path(defaultLanguage, defaultLanguage, page)})
}

func Path(lang, defaultLanguage, page string) string {
	if page == "" && lang == defaultLanguage {
		return "/"
	}

	return "/" + lang + page
}
