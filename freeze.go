package freeze

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/kortschak/utter"
	"github.com/ptdewey/freeze/internal/diff"
	"github.com/ptdewey/freeze/internal/files"
	"github.com/ptdewey/freeze/internal/pretty"
	"github.com/ptdewey/freeze/internal/review"
	"github.com/ptdewey/freeze/internal/transform"
)

const version = "0.1.0"

// TODO: probably make this (and other things) configurable
func init() {
	utter.Config.ElideType = true
	utter.Config.SortKeys = true
}

func SnapString(t testingT, title string, content string) {
	t.Helper()
	SnapStringWithOptions(t, title, content, nil)
}

// SnapStringWithOptions takes a string and applies scrubbers before snapshotting.
func SnapStringWithOptions(t testingT, title string, content string, opts []SnapshotOption) {
	t.Helper()
	config := newSnapshotConfig(opts)

	// Apply scrubbers to the content
	scrubbedContent := transform.ApplyScrubbers(content, adaptScrubbers(config.Scrubbers))

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
		Scrubbers: adaptScrubbers(config.Scrubbers),
		Ignore:    adaptIgnorePatterns(config.Ignore),
	}

	transformedJSON, err := transform.TransformJSON(jsonStr, transformConfig)
	if err != nil {
		t.Error("failed to transform JSON:", err)
		return
	}

	snap(t, title, transformedJSON)
}

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
	scrubbedContent := transform.ApplyScrubbers(content, adaptScrubbers(config.Scrubbers))

	snap(t, title, scrubbedContent)
}

func snap(t testingT, title string, content string) {
	t.Helper()
	testName := t.Name()

	// Capture the caller's file name by walking up the call stack
	// to find the first file that's not freeze.go  TODO: does this actually work for all cases?
	fileName := "unknown"
	for i := 1; i < 10; i++ {
		_, file, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		baseName := filepath.Base(file)
		// Skip frames within freeze.go to get to the actual test file
		if baseName != "freeze.go" {
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

		diffLines := convertDiffLines(diff.Histogram(accepted.Content, snapshot.Content))
		fmt.Println(pretty.DiffSnapshotBox(accepted, snapshot, diffLines))
		t.Error("snapshot mismatch - run 'freeze review' to update")
		return
	}

	if err := files.SaveSnapshot(snapshot, "new"); err != nil {
		t.Error("failed to save snapshot:", err)
		return
	}

	fmt.Println(pretty.NewSnapshotBox(snapshot))
	t.Error("new snapshot created - run 'freeze review' to accept")
}

func convertDiffLines(diffLines []diff.DiffLine) []pretty.DiffLine {
	result := make([]pretty.DiffLine, len(diffLines))
	for i, dl := range diffLines {
		result[i] = pretty.DiffLine{
			OldNumber: dl.OldNumber,
			NewNumber: dl.NewNumber,
			Line:      dl.Line,
			Kind:      pretty.DiffKind(dl.Kind),
		}
	}
	return result
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

// DOCS:
func Review() error {
	return review.Review()
}

func AcceptAll() error {
	return review.AcceptAll()
}

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

// Adapter types to bridge freeze package types with transform package types

type scrubberAdapter struct {
	scrubber Scrubber
}

func (s *scrubberAdapter) Scrub(content string) string {
	return s.scrubber.Scrub(content)
}

func adaptScrubbers(scrubbers []Scrubber) []transform.Scrubber {
	result := make([]transform.Scrubber, len(scrubbers))
	for i, s := range scrubbers {
		result[i] = &scrubberAdapter{scrubber: s}
	}
	return result
}

type ignorePatternAdapter struct {
	pattern IgnorePattern
}

func (i *ignorePatternAdapter) ShouldIgnore(key, value string) bool {
	return i.pattern.ShouldIgnore(key, value)
}

func adaptIgnorePatterns(patterns []IgnorePattern) []transform.IgnorePattern {
	result := make([]transform.IgnorePattern, len(patterns))
	for i, p := range patterns {
		result[i] = &ignorePatternAdapter{pattern: p}
	}
	return result
}
