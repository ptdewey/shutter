package shutter

import (
	"github.com/ptdewey/shutter/internal/diff"
	"github.com/ptdewey/shutter/internal/files"
	"github.com/ptdewey/shutter/internal/pretty"
)

// Snapshot represents a captured test snapshot with metadata.
type Snapshot = files.Snapshot

// DiffLine represents a line in a diff comparison.
type DiffLine = diff.DiffLine

const (
	// DiffShared indicates a line that is unchanged in both versions.
	DiffShared = diff.DiffShared
	// DiffOld indicates a line that was removed.
	DiffOld = diff.DiffOld
	// DiffNew indicates a line that was added.
	DiffNew = diff.DiffNew
)

// Deserialize parses a raw snapshot file string into a Snapshot struct.
func Deserialize(raw string) (*Snapshot, error) {
	return files.Deserialize(raw)
}

// SaveSnapshot writes a snapshot to disk with the specified state ("new" or "accepted").
func SaveSnapshot(snap *Snapshot, state string) error {
	return files.SaveSnapshot(snap, state)
}

// ReadSnapshot reads a snapshot from disk for the given test name and state.
func ReadSnapshot(testName string, state string) (*Snapshot, error) {
	return files.ReadSnapshot(testName, state)
}

// SnapshotFileName returns the snapshot file name for a given test name.
func SnapshotFileName(testName string) string {
	return files.SnapshotFileName(testName)
}

// Histogram computes a line-by-line diff between two strings using the histogram algorithm.
func Histogram(old, new string) []DiffLine {
	return diff.Histogram(old, new)
}

// NewSnapshotBox formats a new snapshot as a pretty-printed box for display.
func NewSnapshotBox(snap *Snapshot) string {
	return pretty.NewSnapshotBox(snap)
}

// DiffSnapshotBox formats a diff between old and new snapshots as a pretty-printed box.
func DiffSnapshotBox(oldSnap, newSnap *Snapshot) string {
	diffLines := diff.Histogram(oldSnap.Content, newSnap.Content)
	return pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)
}
