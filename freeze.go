package freeze

import (
	"fmt"

	"github.com/kortschak/utter"
	"github.com/ptdewey/freeze/internal/diff"
	"github.com/ptdewey/freeze/internal/files"
	"github.com/ptdewey/freeze/internal/pretty"
	"github.com/ptdewey/freeze/internal/review"
)

const version = "0.1.0"

// TODO: probably make this (and other things) configurable
func init() {
	utter.Config.ElideType = true
}

func SnapString(t testingT, title string, content string) {
	t.Helper()
	snap(t, title, content)
}

func Snap(t testingT, title string, values ...any) {
	t.Helper()
	content := formatValues(values...)
	snap(t, title, content)
}

func snap(t testingT, title string, content string) {
	t.Helper()
	testName := t.Name()
	snapWithTitle(t, title, testName, content)
}

func snapWithTitle(t testingT, title string, testName string, content string) {
	t.Helper()

	snapshot := &files.Snapshot{
		Title:   title,
		Name:    testName,
		Content: content,
		Version: version,
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
