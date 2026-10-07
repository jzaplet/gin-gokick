package tracking

import (
	"slices"
	"testing"

	"gokick/app/internal/tracking/events"
)

func emptyEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{"UMAMI_WEBSITE_ID", "GA4_MEASUREMENT_ID", "GOOGLE_ADS_ID", "META_PIXEL_ID", "GOOGLE_ADS_CONVERSIONS"} {
		t.Setenv(key, "")
	}
}

func TestEmptyIDsTurnEveryToolOff(t *testing.T) {
	emptyEnv(t)

	if cfg, err := Read(); err != nil || cfg != (Config{}) {
		t.Errorf("got %+v, %v", cfg, err)
	}
}

func TestReadsTheIDs(t *testing.T) {
	t.Setenv("UMAMI_WEBSITE_ID", "0192f3a4-5b6c-7d8e-9f01-23456789abcd")
	t.Setenv("GA4_MEASUREMENT_ID", "G-AB12CD34EF")
	t.Setenv("GOOGLE_ADS_ID", "AW-123456789")
	t.Setenv("META_PIXEL_ID", "1234567890123456")
	t.Setenv("GOOGLE_ADS_CONVERSIONS", "sign_up=AbC-D_efG")

	cfg, err := Read()
	if err != nil {
		t.Fatal(err)
	}

	want := Config{
		UmamiWebsiteID: "0192f3a4-5b6c-7d8e-9f01-23456789abcd",

		GA4MeasurementID: "G-AB12CD34EF",

		GoogleAdsID: "AW-123456789",

		MetaPixelID: "1234567890123456",

		GoogleAdsConversions: "sign_up=AbC-D_efG",
	}
	if cfg != want {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
}

func TestRejectsInvalidIDs(t *testing.T) {
	tests := []struct {
		name, key, value string
	}{
		{"umami ID in capitals", "UMAMI_WEBSITE_ID", "0192F3A4-5B6C-7D8E-9F01-23456789ABCD"},
		{"umami ID without dashes", "UMAMI_WEBSITE_ID", "0192f3a45b6c7d8e9f0123456789abcd"},
		{"umami ID with a quote", "UMAMI_WEBSITE_ID", `0192f3a4-5b6c-7d8e-9f01-23456789abcd"`},
		{"universal analytics ID", "GA4_MEASUREMENT_ID", "UA-12345-1"},
		{"GA4 ID with a space", "GA4_MEASUREMENT_ID", "G-AB12 CD34"},
		{"ads ID without its prefix", "GOOGLE_ADS_ID", "123456789"},
		{"ads conversion label", "GOOGLE_ADS_ID", "AW-123456789/AbC"},
		{"pixel ID with letters", "META_PIXEL_ID", "12345abc"},
		{"pixel ID with a script", "META_PIXEL_ID", "1234567890<script>"},
		{"pixel ID with a leading zero", "META_PIXEL_ID", "0123456789012345"},
		{"unknown conversion", "GOOGLE_ADS_CONVERSIONS", "sign_in=AbC"},
		{"conversion without a label", "GOOGLE_ADS_CONVERSIONS", "sign_up="},
		{"label with a slash", "GOOGLE_ADS_CONVERSIONS", "sign_up=AW-1/AbC"},
		{"conversion twice", "GOOGLE_ADS_CONVERSIONS", "sign_up=AbC,sign_up=XyZ"},
		{"trailing comma", "GOOGLE_ADS_CONVERSIONS", "sign_up=AbC,"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emptyEnv(t)
			t.Setenv(tt.key, tt.value)

			if _, err := Read(); err == nil {
				t.Errorf("%s=%q passed", tt.key, tt.value)
			}
		})
	}
}

func TestOnlyToolsWithCookiesNeedConsent(t *testing.T) {
	if (&Config{UmamiWebsiteID: "0192f3a4-5b6c-7d8e-9f01-23456789abcd"}).NeedsConsent() {
		t.Error("umami needs consent")
	}

	for _, cfg := range []*Config{{GA4MeasurementID: "G-AB12CD34EF"}, {GoogleAdsID: "AW-123456789"}, {MetaPixelID: "1234567890123456"}} {
		if cfg.NeedsConsent() == false {
			t.Errorf("%+v needs no consent", cfg)
		}
	}
}

func TestListsConversionsWithoutALabelOnlyWithGoogleAds(t *testing.T) {
	for name, tt := range map[string]struct {
		tools Config

		missing []events.AdsConversion
	}{
		"without google ads": {Config{GoogleAdsConversions: ""}, nil},

		"with a label": {Config{GoogleAdsID: "AW-123456789", GoogleAdsConversions: "sign_up=AbC"}, nil},

		"without a label": {Config{GoogleAdsID: "AW-123456789"}, []events.AdsConversion{events.AdsConversionSignUp}},
	} {
		t.Run(name, func(t *testing.T) {
			if got := tt.tools.UnlabeledConversions(); slices.Equal(got, tt.missing) == false {
				t.Errorf("got %v, want %v", got, tt.missing)
			}
		})
	}
}
