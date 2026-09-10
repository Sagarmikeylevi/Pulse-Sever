package shared

import "time"

// Legacy IANA timezone aliases mapped to their canonical names.
// Source: https://github.com/eggert/tz/blob/main/backward
var timezoneAliases = map[string]string{
	// Asia
	"Asia/Calcutta":    "Asia/Kolkata",
	"Asia/Saigon":      "Asia/Ho_Chi_Minh",
	"Asia/Katmandu":    "Asia/Kathmandu",
	"Asia/Rangoon":     "Asia/Yangon",
	"Asia/Thimbu":      "Asia/Thimphu",
	"Asia/Ujung_Pandang": "Asia/Makassar",
	"Asia/Ulan_Bator":  "Asia/Ulaanbaatar",
	"Asia/Dacca":       "Asia/Dhaka",
	"Asia/Macao":       "Asia/Macau",

	// US
	"US/Eastern":  "America/New_York",
	"US/Central":  "America/Chicago",
	"US/Mountain": "America/Denver",
	"US/Pacific":  "America/Los_Angeles",
	"US/Hawaii":   "Pacific/Honolulu",
	"US/Alaska":   "America/Anchorage",
	"US/Arizona":  "America/Phoenix",

	// Europe
	"Europe/Kiev": "Europe/Kyiv",

	// Atlantic
	"Atlantic/Faeroe": "Atlantic/Faroe",

	// Pacific
	"Pacific/Samoa":   "Pacific/Pago_Pago",
	"Pacific/Ponape":  "Pacific/Pohnpei",
	"Pacific/Truk":    "Pacific/Chuuk",

	// Other
	"Australia/ACT":       "Australia/Sydney",
	"Australia/Queensland": "Australia/Brisbane",
	"Canada/Eastern":      "America/Toronto",
	"Canada/Central":      "America/Winnipeg",
	"Canada/Pacific":      "America/Vancouver",
}

// NormalizeTimezone validates the timezone and returns the canonical IANA name.
// Returns the canonical name and nil error if valid, or empty string and error if invalid.
func NormalizeTimezone(tz string) (string, error) {
	_, err := time.LoadLocation(tz)
	if err != nil {
		return "", err
	}

	if canonical, ok := timezoneAliases[tz]; ok {
		return canonical, nil
	}

	return tz, nil
}
