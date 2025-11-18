package api

import (
	"github.com/ptdewey/freeze/internal/diff"
	"github.com/ptdewey/freeze/internal/files"
	"github.com/ptdewey/freeze/internal/pretty"
)

type Snapshot = files.Snapshot

type DiffLine = diff.DiffLine

const (
	DiffShared = diff.DiffShared
	DiffOld    = diff.DiffOld
	DiffNew    = diff.DiffNew
)

func Deserialize(raw string) (*Snapshot, error) {
	return files.Deserialize(raw)
}

func SaveSnapshot(snap *Snapshot, state string) error {
	return files.SaveSnapshot(snap, state)
}

func ReadSnapshot(testName string, state string) (*Snapshot, error) {
	return files.ReadSnapshot(testName, state)
}

func SnapshotFileName(testName string) string {
	return files.SnapshotFileName(testName)
}

func Histogram(old, new string) []DiffLine {
	return diff.Histogram(old, new)
}

func NewSnapshotBox(snap *Snapshot) string {
	return pretty.NewSnapshotBox(snap)
}

func DiffSnapshotBox(old, new *Snapshot) string {
	diffLines := convertDiffLines(diff.Histogram(old.Content, new.Content))
	return pretty.DiffSnapshotBox(old, new, diffLines)
}

func convertDiffLines(diffLines []diff.DiffLine) []pretty.DiffLine {
	result := make([]pretty.DiffLine, len(diffLines))
	for i, dl := range diffLines {
		result[i] = pretty.DiffLine{
			Number: dl.Number,
			Line:   dl.Line,
			Kind:   pretty.DiffKind(dl.Kind),
		}
	}
	return result
}

func Red(s string) string {
	return pretty.Red(s)
}

func Green(s string) string {
	return pretty.Green(s)
}

func Yellow(s string) string {
	return pretty.Yellow(s)
}

func Blue(s string) string {
	return pretty.Blue(s)
}

func Gray(s string) string {
	return pretty.Gray(s)
}

func Bold(s string) string {
	return pretty.Bold(s)
}

func TerminalWidth() int {
	return pretty.TerminalWidth()
}

func ClearScreen() {
	pretty.ClearScreen()
}

func ClearLine() {
	pretty.ClearLine()
}
