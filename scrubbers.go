package shutter

import (
	"regexp"
	"strings"
)

// regexScrubber replaces all matches of a regex pattern with a replacement string.
type regexScrubber struct {
	pattern     *regexp.Regexp
	replacement string
}

func (r *regexScrubber) isOption() {}

func (r *regexScrubber) Scrub(content string) string {
	return r.pattern.ReplaceAllString(content, r.replacement)
}

// ScrubRegex creates a scrubber that replaces all matches of the given
// regex pattern with the replacement string.
//
// Example:
//
//	shutter.ScrubRegex(`user-\d+`, "<USER_ID>")
func ScrubRegex(pattern string, replacement string) Scrubber {
	re := regexp.MustCompile(pattern)
	return &regexScrubber{
		pattern:     re,
		replacement: replacement,
	}
}

// exactMatchScrubber replaces exact string matches with a replacement.
type exactMatchScrubber struct {
	match       string
	replacement string
}

func (e *exactMatchScrubber) isOption() {}

func (e *exactMatchScrubber) Scrub(content string) string {
	return strings.ReplaceAll(content, e.match, e.replacement)
}

// ScrubExact creates a scrubber that replaces exact string matches.
//
// Example:
//
//	shutter.ScrubExact("secret_value", "<REDACTED>")
func ScrubExact(match string, replacement string) Scrubber {
	return &exactMatchScrubber{
		match:       match,
		replacement: replacement,
	}
}

// Common regex patterns for scrubbing
var (
	uuidPattern    = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	iso8601Pattern = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})?`)
	emailPattern   = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	// Unix timestamp pattern - matches 10-13 digit numbers (Unix timestamps in seconds or milliseconds)
	// Note: This is aggressive and may match other numbers. Use with caution or customize.
	unixTsPattern = regexp.MustCompile(`\b\d{10,13}\b`)
	// IPv4 pattern with basic range validation
	ipv4Pattern = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
	// Credit card pattern - matches 16 digit numbers with optional separators
	creditCardPattern = regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`)
	jwtPattern        = regexp.MustCompile(`eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*`)
	// Date patterns
	datePattern = regexp.MustCompile(`\b\d{4}[-/]\d{2}[-/]\d{2}\b|\b\d{2}[-/]\d{2}[-/]\d{4}\b`)
	// API key pattern - matches patterns like: sk_live_..., pk_test_..., api_key_...
	apiKeyPattern = regexp.MustCompile(`\b(sk|pk|api[_-]?key)[_-](live|test|prod|dev)[_-][a-zA-Z0-9]+\b`)
)

// ScrubUUID replaces all UUIDs with "<UUID>".
//
// Example:
//
//	shutter.Snap(t, "user", user, shutter.ScrubUUID())
func ScrubUUID() Scrubber {
	return &regexScrubber{
		pattern:     uuidPattern,
		replacement: "<UUID>",
	}
}

// ScrubTimestamp replaces ISO8601 timestamps with "<TIMESTAMP>".
//
// Example:
//
//	shutter.Snap(t, "event", event, shutter.ScrubTimestamp())
func ScrubTimestamp() Scrubber {
	return &regexScrubber{
		pattern:     iso8601Pattern,
		replacement: "<TIMESTAMP>",
	}
}

// ScrubEmail replaces email addresses with "<EMAIL>".
//
// Example:
//
//	shutter.Snap(t, "user", user, shutter.ScrubEmail())
func ScrubEmail() Scrubber {
	return &regexScrubber{
		pattern:     emailPattern,
		replacement: "<EMAIL>",
	}
}

// ScrubUnixTimestamp replaces Unix timestamps (10-13 digits) with "<UNIX_TS>".
// Note: This is aggressive and may match other long numbers. For more conservative
// scrubbing with context keywords, use ScrubRegex with a custom pattern.
//
// Example:
//
//	shutter.Snap(t, "data", data, shutter.ScrubUnixTimestamp())
func ScrubUnixTimestamp() Scrubber {
	return &regexScrubber{
		pattern:     unixTsPattern,
		replacement: "<UNIX_TS>",
	}
}

// ScrubIP replaces IPv4 addresses with "<IP>".
//
// Example:
//
//	shutter.Snap(t, "request", request, shutter.ScrubIP())
func ScrubIP() Scrubber {
	return &regexScrubber{
		pattern:     ipv4Pattern,
		replacement: "<IP>",
	}
}

// ScrubCreditCard replaces credit card numbers with "<CREDIT_CARD>".
//
// Example:
//
//	shutter.Snap(t, "payment", payment, shutter.ScrubCreditCard())
func ScrubCreditCard() Scrubber {
	return &regexScrubber{
		pattern:     creditCardPattern,
		replacement: "<CREDIT_CARD>",
	}
}

// ScrubJWT replaces JWT tokens with "<JWT>".
//
// Example:
//
//	shutter.Snap(t, "auth", authData, shutter.ScrubJWT())
func ScrubJWT() Scrubber {
	return &regexScrubber{
		pattern:     jwtPattern,
		replacement: "<JWT>",
	}
}

// ScrubDate replaces various date formats with "<DATE>".
//
// Example:
//
//	shutter.Snap(t, "data", data, shutter.ScrubDate())
func ScrubDate() Scrubber {
	return &regexScrubber{
		pattern:     datePattern,
		replacement: "<DATE>",
	}
}

// ScrubAPIKey replaces common API key patterns with "<API_KEY>".
// Matches patterns like: sk_live_..., pk_test_..., api_key_...
//
// Example:
//
//	shutter.Snap(t, "config", config, shutter.ScrubAPIKey())
func ScrubAPIKey() Scrubber {
	return &regexScrubber{
		pattern:     apiKeyPattern,
		replacement: "<API_KEY>",
	}
}

// customScrubber allows users to provide a custom scrubbing function.
type customScrubber struct {
	scrubFunc func(string) string
}

func (c *customScrubber) isOption() {}

func (c *customScrubber) Scrub(content string) string {
	return c.scrubFunc(content)
}

// ScrubWith creates a scrubber using a custom function.
// The function receives the snapshot content and should return the scrubbed content.
//
// Example:
//
//	shutter.Snap(t, "data", data,
//	    shutter.ScrubWith(func(content string) string {
//	        return strings.ReplaceAll(content, "localhost", "<HOST>")
//	    }),
//	)
func ScrubWith(scrubFunc func(string) string) Scrubber {
	return &customScrubber{
		scrubFunc: scrubFunc,
	}
}
