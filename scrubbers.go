package freeze

import (
	"regexp"
	"strings"
)

// Scrubber transforms content before snapshotting, typically to remove
// or replace dynamic or sensitive data.
type Scrubber interface {
	Scrub(content string) string
}

// regexScrubber replaces all matches of a regex pattern with a replacement string.
type regexScrubber struct {
	pattern     *regexp.Regexp
	replacement string
}

func (r *regexScrubber) Scrub(content string) string {
	return r.pattern.ReplaceAllString(content, r.replacement)
}

// RegexScrubber creates a scrubber that replaces all matches of the given
// regex pattern with the replacement string.
func RegexScrubber(pattern string, replacement string) SnapshotOption {
	re := regexp.MustCompile(pattern)
	return WithScrubber(&regexScrubber{
		pattern:     re,
		replacement: replacement,
	})
}

// exactMatchScrubber replaces exact string matches with a replacement.
type exactMatchScrubber struct {
	match       string
	replacement string
}

func (e *exactMatchScrubber) Scrub(content string) string {
	return strings.ReplaceAll(content, e.match, e.replacement)
}

// ExactMatchScrubber creates a scrubber that replaces exact string matches.
func ExactMatchScrubber(match string, replacement string) SnapshotOption {
	return WithScrubber(&exactMatchScrubber{
		match:       match,
		replacement: replacement,
	})
}

// Common regex patterns for scrubbing
// TODO: review these
var (
	uuidPattern       = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	iso8601Pattern    = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})?`)
	emailPattern      = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	unixTsPattern     = regexp.MustCompile(`\b\d{10,13}\b`)
	ipv4Pattern       = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	creditCardPattern = regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`)
	jwtPattern        = regexp.MustCompile(`eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*`)
)

// ScrubUUIDs replaces all UUIDs with "<UUID>".
func ScrubUUIDs() SnapshotOption {
	return WithScrubber(&regexScrubber{
		pattern:     uuidPattern,
		replacement: "<UUID>",
	})
}

// ScrubTimestamps replaces ISO8601 timestamps with "<TIMESTAMP>".
func ScrubTimestamps() SnapshotOption {
	return WithScrubber(&regexScrubber{
		pattern:     iso8601Pattern,
		replacement: "<TIMESTAMP>",
	})
}

// ScrubEmails replaces email addresses with "<EMAIL>".
func ScrubEmails() SnapshotOption {
	return WithScrubber(&regexScrubber{
		pattern:     emailPattern,
		replacement: "<EMAIL>",
	})
}

// ScrubUnixTimestamps replaces Unix timestamps (10-13 digits) with "<UNIX_TS>".
func ScrubUnixTimestamps() SnapshotOption {
	return WithScrubber(&regexScrubber{
		pattern:     unixTsPattern,
		replacement: "<UNIX_TS>",
	})
}

// ScrubIPAddresses replaces IPv4 addresses with "<IP>".
func ScrubIPAddresses() SnapshotOption {
	return WithScrubber(&regexScrubber{
		pattern:     ipv4Pattern,
		replacement: "<IP>",
	})
}

// ScrubCreditCards replaces credit card numbers with "<CREDIT_CARD>".
func ScrubCreditCards() SnapshotOption {
	return WithScrubber(&regexScrubber{
		pattern:     creditCardPattern,
		replacement: "<CREDIT_CARD>",
	})
}

// ScrubJWTs replaces JWT tokens with "<JWT>".
func ScrubJWTs() SnapshotOption {
	return WithScrubber(&regexScrubber{
		pattern:     jwtPattern,
		replacement: "<JWT>",
	})
}

// ScrubDates replaces dates in various formats with "<DATE>".
// Matches formats like: 2023-01-15, 01/15/2023, 15-01-2023
func ScrubDates() SnapshotOption {
	datePattern := regexp.MustCompile(`\b\d{4}[-/]\d{2}[-/]\d{2}\b|\b\d{2}[-/]\d{2}[-/]\d{4}\b`)
	return WithScrubber(&regexScrubber{
		pattern:     datePattern,
		replacement: "<DATE>",
	})
}

// ScrubAPIKeys replaces common API key patterns with "<API_KEY>".
// Matches patterns like: sk_live_..., pk_test_..., api_key_...
func ScrubAPIKeys() SnapshotOption {
	apiKeyPattern := regexp.MustCompile(`\b(sk|pk|api[_-]?key)[_-](live|test|prod|dev)[_-][a-zA-Z0-9]+\b`)
	return WithScrubber(&regexScrubber{
		pattern:     apiKeyPattern,
		replacement: "<API_KEY>",
	})
}

type customScrubber struct {
	scrubFunc func(string) string
}

func (c *customScrubber) Scrub(content string) string {
	return c.scrubFunc(content)
}

// CustomScrubber creates a scrubber using a custom function.
func CustomScrubber(scrubFunc func(string) string) SnapshotOption {
	return WithScrubber(&customScrubber{
		scrubFunc: scrubFunc,
	})
}
