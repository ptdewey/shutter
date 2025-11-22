package shutter

import (
	"fmt"

	"github.com/kortschak/utter"
	"github.com/ptdewey/shutter/internal/review"
	"github.com/ptdewey/shutter/internal/snapshots"
	"github.com/ptdewey/shutter/internal/transform"
)

// snapshotFormatVersion indicates the snapshot format version used by this library.
// This is automatically included in snapshot metadata for compatibility checking
// when the snapshot format changes in future versions.
const snapshotFormatVersion = "0.1.0"

// utterConfig is a configured instance of utter for consistent formatting.
// This avoids modifying global state and ensures snapshot formatting is isolated.
var utterConfig = &utter.ConfigState{
	Indent:    "  ",
	ElideType: true,
	SortKeys:  true,
}

// Option is a marker interface for all snapshot options.
// This allows compile-time type safety while supporting different option types.
type Option interface {
	isOption()
}

// Scrubber transforms content before snapshotting, typically to replace
// dynamic or sensitive data with stable placeholders.
//
// Scrubbers are applied in the order they are provided. Later scrubbers
// can transform the output of earlier scrubbers.
//
// Example:
//
//	shutter.Snap(t, "user data", user,
//	    shutter.ScrubUUID(),        // First: UUIDs -> <UUID>
//	    shutter.ScrubEmail(),       // Second: emails -> <EMAIL>
//	)
type Scrubber interface {
	Option
	Scrub(content string) string
}

// IgnorePattern determines whether a key-value pair should be excluded
// from JSON snapshots. This is useful for removing fields that change
// frequently or contain sensitive data.
//
// IgnorePatterns only work with SnapJSON. Using them with Snap or SnapString
// will result in an error.
type IgnorePattern interface {
	Option
	ShouldIgnore(key, value string) bool
}

// Snap takes a single value, formats it, and creates a snapshot with the given title.
// Complex types are formatted using a pretty-printer for readability.
//
// Options can be provided to scrub sensitive or dynamic data before snapshotting.
// Only Scrubber options are supported; IgnorePattern options will cause an error.
//
// Example:
//
//	user := User{ID: "123", Email: "user@example.com"}
//	shutter.Snap(t, "user data", user,
//	    shutter.ScrubUUID(),
//	    shutter.ScrubEmail(),
//	)
func Snap(t snapshots.T, title string, value any, opts ...Option) {
	t.Helper()

	scrubbers, ignores := separateOptions(opts)

	if len(ignores) > 0 {
		t.Error(fmt.Sprintf("snapshot %q: IgnorePattern options are not supported with Snap; use SnapJSON instead", title))
		return
	}

	content := formatValue(value)
	scrubbedContent := applyScrubbers(content, scrubbers)

	snapshots.Snap(t, title, snapshotFormatVersion, scrubbedContent)
}

// SnapMany takes multiple values, formats them, and creates a snapshot with the given title.
// This is useful when you want to snapshot multiple related values together.
//
// Options can be provided to scrub sensitive or dynamic data before snapshotting.
// Only Scrubber options are supported; IgnorePattern options will cause an error.
//
// Example:
//
//	shutter.SnapMany(t, "request and response",
//	    []any{request, response},
//	    shutter.ScrubUUID(),
//	    shutter.ScrubTimestamp(),
//	)
func SnapMany(t snapshots.T, title string, values []any, opts ...Option) {
	t.Helper()

	scrubbers, ignores := separateOptions(opts)

	if len(ignores) > 0 {
		t.Error(fmt.Sprintf("snapshot %q: IgnorePattern options are not supported with SnapMany; use SnapJSON instead", title))
		return
	}

	content := formatValues(values...)
	scrubbedContent := applyScrubbers(content, scrubbers)

	snapshots.Snap(t, title, snapshotFormatVersion, scrubbedContent)
}

// SnapString takes a string value and creates a snapshot with the given title.
// This is useful for snapshotting generated text, logs, or other string content.
//
// Options can be provided to scrub sensitive or dynamic data before snapshotting.
// Only Scrubber options are supported; IgnorePattern options will cause an error.
//
// Example:
//
//	output := generateReport()
//	shutter.SnapString(t, "report output", output,
//	    shutter.ScrubTimestamp(),
//	)
func SnapString(t snapshots.T, title string, content string, opts ...Option) {
	t.Helper()

	scrubbers, ignores := separateOptions(opts)

	if len(ignores) > 0 {
		t.Error(fmt.Sprintf("snapshot %q: IgnorePattern options are not supported with SnapString; use SnapJSON instead", title))
		return
	}

	scrubbedContent := applyScrubbers(content, scrubbers)

	snapshots.Snap(t, title, snapshotFormatVersion, scrubbedContent)
}

// SnapJSON takes a JSON string, validates it, and pretty-prints it with
// consistent formatting before snapshotting. This preserves the raw JSON
// format while ensuring valid JSON structure.
//
// Options can be provided to apply both Scrubbers and IgnorePatterns.
// IgnorePatterns remove fields from the JSON structure before scrubbing.
// Scrubbers then transform the remaining content.
//
// Example:
//
//	jsonStr := `{"id": "550e8400-...", "email": "user@example.com", "password": "secret"}`
//	shutter.SnapJSON(t, "user response", jsonStr,
//	    shutter.IgnoreKey("password"),     // First: remove password field
//	    shutter.ScrubUUID(),               // Second: scrub remaining UUIDs
//	    shutter.ScrubEmail(),              // Third: scrub emails
//	)
func SnapJSON(t snapshots.T, title string, jsonStr string, opts ...Option) {
	t.Helper()

	scrubbers, ignores := separateOptions(opts)

	// Transform the JSON with ignore patterns and scrubbers
	transformConfig := &transform.Config{
		Scrubbers: toTransformScrubbers(scrubbers),
		Ignore:    toTransformIgnorePatterns(ignores),
	}

	transformedJSON, err := transform.TransformJSON(jsonStr, transformConfig)
	if err != nil {
		t.Error(fmt.Sprintf("snapshot %q: failed to transform JSON: %v", title, err))
		return
	}

	snapshots.Snap(t, title, snapshotFormatVersion, transformedJSON)
}

// Review launches an interactive review session to accept or reject snapshot changes.
func Review() error {
	return review.Review()
}

// AcceptAll accepts all pending snapshot changes without review.
func AcceptAll() error {
	return review.AcceptAll()
}

// RejectAll rejects all pending snapshot changes without review.
func RejectAll() error {
	return review.RejectAll()
}

// formatValue formats a single value using the configured utter instance.
func formatValue(v any) string {
	return utterConfig.Sdump(v)
}

// formatValues formats multiple values using the configured utter instance.
func formatValues(values ...any) string {
	var result string
	for _, v := range values {
		result += formatValue(v)
	}
	return result
}

// separateOptions splits options into scrubbers and ignore patterns.
func separateOptions(opts []Option) (scrubbers []Scrubber, ignores []IgnorePattern) {
	for _, opt := range opts {
		switch o := opt.(type) {
		case IgnorePattern:
			ignores = append(ignores, o)
		case Scrubber:
			scrubbers = append(scrubbers, o)
		default:
			// This shouldn't happen if Option interface is properly implemented
			panic(fmt.Sprintf("unknown option type: %T", opt))
		}
	}
	return scrubbers, ignores
}

// applyScrubbers applies all scrubbers to content in sequence.
func applyScrubbers(content string, scrubbers []Scrubber) string {
	for _, scrubber := range scrubbers {
		content = scrubber.Scrub(content)
	}
	return content
}

// scrubberAdapter adapts a Scrubber to the transform.Scrubber interface.
type scrubberAdapter struct {
	scrubber Scrubber
}

func (s *scrubberAdapter) Scrub(content string) string {
	return s.scrubber.Scrub(content)
}

func toTransformScrubbers(scrubbers []Scrubber) []transform.Scrubber {
	result := make([]transform.Scrubber, len(scrubbers))
	for i, scrubber := range scrubbers {
		result[i] = &scrubberAdapter{scrubber: scrubber}
	}
	return result
}

// ignoreAdapter adapts an IgnorePattern to the transform.IgnorePattern interface.
type ignoreAdapter struct {
	ignore IgnorePattern
}

func (i *ignoreAdapter) ShouldIgnore(key, value string) bool {
	return i.ignore.ShouldIgnore(key, value)
}

func toTransformIgnorePatterns(ignores []IgnorePattern) []transform.IgnorePattern {
	result := make([]transform.IgnorePattern, len(ignores))
	for i, ignore := range ignores {
		result[i] = &ignoreAdapter{ignore: ignore}
	}
	return result
}
