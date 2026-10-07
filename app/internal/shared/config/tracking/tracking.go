package tracking

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"gokick/app/internal/tracking/events"
)

var uuidID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

var ga4ID = regexp.MustCompile(`^G-[A-Z0-9]+$`)

var adsID = regexp.MustCompile(`^AW-\d+$`)

var pixelID = regexp.MustCompile(`^[1-9]\d{0,25}$`)

var conversionLabel = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type Config struct {
	UmamiWebsiteID string

	GA4MeasurementID string

	GoogleAdsID string

	MetaPixelID string

	GoogleAdsConversions string
}

type trackingID struct {
	key string

	pattern *regexp.Regexp

	example string

	target *string
}

func Read() (Config, error) {
	var cfg Config

	ids := []trackingID{
		{"UMAMI_WEBSITE_ID", uuidID, "a lowercase UUID", &cfg.UmamiWebsiteID},
		{"GA4_MEASUREMENT_ID", ga4ID, "G-XXXXXXXXXX", &cfg.GA4MeasurementID},
		{"GOOGLE_ADS_ID", adsID, "AW-123456789", &cfg.GoogleAdsID},
		{"META_PIXEL_ID", pixelID, "digits without a leading zero", &cfg.MetaPixelID},
	}
	for _, id := range ids {
		value := os.Getenv(id.key)
		if value == "" || id.pattern.MatchString(value) {
			*id.target = value

			continue
		}

		return Config{}, fmt.Errorf("%s must be %s or empty, got %q", id.key, id.example, value)
	}

	cfg.GoogleAdsConversions = os.Getenv("GOOGLE_ADS_CONVERSIONS")
	if _, err := conversionLabels(cfg.GoogleAdsConversions); err != nil {
		return Config{}, fmt.Errorf("GOOGLE_ADS_CONVERSIONS must be conversion=label pairs split by commas: %w", err)
	}

	return cfg, nil
}

func (c *Config) NeedsConsent() bool {
	return c.GA4MeasurementID != "" || c.GoogleAdsID != "" || c.MetaPixelID != ""
}

func (c *Config) UnlabeledConversions() []events.AdsConversion {
	if c.GoogleAdsID == "" {
		return nil
	}

	labels, _ := conversionLabels(c.GoogleAdsConversions)
	var missing []events.AdsConversion

	for _, conversion := range events.AdsConversions {
		if _, ok := labels[conversion]; ok == false {
			missing = append(missing, conversion)
		}
	}

	return missing
}

func conversionLabels(value string) (map[events.AdsConversion]string, error) {
	labels := map[events.AdsConversion]string{}
	if value == "" {
		return labels, nil
	}

	for pair := range strings.SplitSeq(value, ",") {
		name, label, _ := strings.Cut(pair, "=")

		conversion := events.AdsConversion(name)
		switch {
		case slices.Contains(events.AdsConversions, conversion) == false:
			return nil, fmt.Errorf("unknown conversion %q", name)
		case conversionLabel.MatchString(label) == false:
			return nil, fmt.Errorf("conversion %q has the label %q", name, label)
		case labels[conversion] != "":
			return nil, fmt.Errorf("conversion %q twice", name)
		}

		labels[conversion] = label
	}

	return labels, nil
}
