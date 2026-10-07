package csp

const googleTagManager = "https://www.googletagmanager.com"

const turnstile = "https://challenges.cloudflare.com"

func GoogleAnalytics() Service {
	return Service{
		ScriptSrc: []string{googleTagManager},

		FrameSrc: []string{googleTagManager},

		TrustedTypes: []string{"goog#html", "'allow-duplicates'"},
	}
}

func GoogleAds() Service {
	return Service{
		ScriptSrc: []string{googleTagManager, "https://www.googleadservices.com", "https://www.google.com"},

		FrameSrc: []string{googleTagManager},

		TrustedTypes: []string{"goog#html", "'allow-duplicates'"},
	}
}

func MetaPixel() Service {
	return Service{
		ScriptSrc: []string{"https://connect.facebook.net"},

		TrustedTypes: []string{"connect.facebook.net/fbevents", "facebook.com/signals/iwl"},
	}
}

func Turnstile() Service {
	return Service{ScriptSrc: []string{turnstile}, FrameSrc: []string{turnstile}}
}

func Vue() Service {
	return Service{TrustedTypes: []string{"vue"}}
}
