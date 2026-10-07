package posix

import "testing"

func TestParseRefusesWhatIsNotALocale(t *testing.T) {
	for name, value := range map[string]string{
		"only a language": "en",

		"a dash": "en-US",

		"capital language": "EN_US",

		"small country": "en_us",

		"a space": " en_US",

		"nothing": "",

		"unknown language": "xx_US",

		"private country": "en_XX",

		"old country code": "en_UK",

		"old language code": "iw_IL",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(value); err == nil {
				t.Errorf("%q passed", value)
			}
		})
	}
}
