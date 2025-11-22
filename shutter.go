package shutter

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/kortschak/utter"
	"github.com/ptdewey/shutter/internal/diff"
	"github.com/ptdewey/shutter/internal/files"
	"github.com/ptdewey/shutter/internal/pretty"
	"github.com/ptdewey/shutter/internal/review"
	"github.com/ptdewey/shutter/internal/transform"
)

const version = "0.1.0"

// TODO: probably make this (and other things) configurable
func init() {
	utter.Config.ElideType = true
	utter.Config.SortKeys = true
}

// SnapString takes a string value and creates a snapshot with the given title.
func SnapString(t testingT, title string, content string) {
	t.Helper()
	SnapStringWithOptions(t, title, content, nil)
}

// SnapStringWithOptions takes a string and applies scrubbers before snapshotting.
func SnapStringWithOptions(t testingT, title string, content string, opts []SnapshotOption) {
	t.Helper()
	config := newSnapshotConfig(opts)

	// Apply scrubbers to the content
	scrubbedContent := transform.ApplyScrubbers(content, toTransformScrubbers(config.Scrubbers))

	snap(t, title, scrubbedContent)
}

// SnapJSON takes a JSON string, validates it, and pretty-prints it with
// consistent formatting before snapshotting. This preserves the raw JSON
// format while ensuring valid JSON structure.
func SnapJSON(t testingT, title string, jsonStr string) {
	t.Helper()
	SnapJSONWithOptions(t, title, jsonStr, nil)
}

// SnapJSONWithOptions takes a JSON string and applies scrubbers and ignore patterns
// before snapshotting. This allows filtering sensitive data and normalizing dynamic values.
func SnapJSONWithOptions(t testingT, title string, jsonStr string, opts []SnapshotOption) {
	t.Helper()

	config := newSnapshotConfig(opts)

	// Transform the JSON with ignore patterns and scrubbers
	transformConfig := &transform.Config{
		Scrubbers: toTransformScrubbers(config.Scrubbers),
		Ignore:    toTransformIgnorePatterns(config.Ignore),
	}

	transformedJSON, err := transform.TransformJSON(jsonStr, transformConfig)
	if err != nil {
		t.Error("failed to transform JSON:", err)
		return
	}

	snap(t, title, transformedJSON)
}

// Snap takes any values, formats them, and creates a snapshot with the given title.
// For complex types, values are formatted using a pretty-printer.
func Snap(t testingT, title string, values ...any) {
	t.Helper()
	SnapWithOptions(t, title, nil, values...)
}

// SnapWithOptions takes any values, formats them, and applies scrubbers before snapshotting.
// For structured data (maps, slices, structs), scrubbers are applied to the formatted output.
func SnapWithOptions(t testingT, title string, opts []SnapshotOption, values ...any) {
	t.Helper()
	config := newSnapshotConfig(opts)

	content := formatValues(values...)

	// Apply scrubbers to the formatted content
	scrubbedContent := transform.ApplyScrubbers(content, toTransformScrubbers(config.Scrubbers))

	snap(t, title, scrubbedContent)
}

func snap(t testingT, title string, content string) {
	t.Helper()
	testName := t.Name()

	// Capture the caller's file name by walking up the call stack
	// to find the first file that's not shutter.go  TODO: does this actually work for all cases?
	fileName := "unknown"
	for i := 1; i < 10; i++ {
		_, file, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		baseName := filepath.Base(file)
		// Skip frames within shutter.go to get to the actual test file
		if baseName != "shutter.go" {
			fileName = baseName
			break
		}
	}

	snapWithTitle(t, title, testName, fileName, content)
}

func snapWithTitle(t testingT, title string, testName string, fileName string, content string) {
	t.Helper()

	snapshot := &files.Snapshot{
		Title:    title,
		Test:     testName,
		FileName: fileName,
		Content:  content,
		Version:  version,
	}

	accepted, err := files.ReadAccepted(testName)
	if err == nil {
		if accepted.Content == content {
			return
		}

		if err := files.SaveSnapshot(snapshot, "new"); err != nil {
			t.Error("failed to save snapshot:", err)
			return
		}

		diffLines := diff.Histogram(accepted.Content, snapshot.Content)
		fmt.Println(pretty.DiffSnapshotBox(accepted, snapshot, diffLines))
		t.Error("snapshot mismatch - run 'shutter review' to update")
		return
	}

	if err := files.SaveSnapshot(snapshot, "new"); err != nil {
		t.Error("failed to save snapshot:", err)
		return
	}

	fmt.Println(pretty.NewSnapshotBox(snapshot))
	t.Error("new snapshot created - run 'shutter review' to accept")
}

func formatValues(values ...any) string {
	var result string
	for _, v := range values {
		result += formatValue(v)
	}
	return result
}

func formatValue(v any) string {
	return utter.Sdump(v)
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

type testingT interface {
	Helper()
	Skip(...any)
	Skipf(string, ...any)
	SkipNow()
	Name() string
	Error(...any)
	Log(...any)
	Cleanup(func())
}

// Type conversion helpers to bridge shutter package types with transform package types.
// These work because the interfaces have identical method signatures (structural typing).

func toTransformScrubbers(scrubbers []Scrubber) []transform.Scrubber {
	result := make([]transform.Scrubber, len(scrubbers))
	for i, s := range scrubbers {
		result[i] = s
	}
	return result
}

func toTransformIgnorePatterns(patterns []IgnorePattern) []transform.IgnorePattern {
	result := make([]transform.IgnorePattern, len(patterns))
	for i, p := range patterns {
		result[i] = p
	}
	return result
}
