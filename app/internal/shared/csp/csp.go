package csp

import (
	"gokick/app/core/csp"
	"gokick/app/internal/shared/config/tracking"
)

const umami = "https://analytics.strategio.dev"

//tsgen:assets/shared/ScriptLoader/types/ScriptURL.ts ScriptURL union
type ScriptURL string

const GoogleTag ScriptURL = "https://www.googletagmanager.com/gtag/js"

const MetaPixel ScriptURL = "https://connect.facebook.net/en_US/fbevents.js"

//tsgen:assets/shared/ScriptLoader/types/TrustedTypesPolicy.ts TrustedTypesPolicy union
type TrustedTypesPolicy string

const ScriptLoader TrustedTypesPolicy = "script-loader"

func Policy(tools *tracking.Config) csp.Policy {
	var services []csp.Service
	if tools.GA4MeasurementID != "" {
		services = append(services, csp.GoogleAnalytics())
	}

	if tools.GoogleAdsID != "" {
		services = append(services, csp.GoogleAds())
	}

	if tools.MetaPixelID != "" {
		services = append(services, csp.MetaPixel())
	}

	if tools.UmamiWebsiteID != "" {
		services = append(services, csp.Service{ScriptSrc: []string{umami}})
	}

	services = append(services, csp.Turnstile(), csp.Vue())
	if tools.NeedsConsent() {
		services = append(services, csp.Service{TrustedTypes: []string{string(ScriptLoader)}})
	}

	return csp.Policy{Services: services}
}
