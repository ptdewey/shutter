package snapshots

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/ptdewey/shutter/internal/diff"
	"github.com/ptdewey/shutter/internal/files"
	"github.com/ptdewey/shutter/internal/pretty"
)

type T interface {
	Helper()
	Skip(...any)
	Skipf(string, ...any)
	SkipNow()
	Name() string
	Error(...any)
	Log(...any)
	Cleanup(func())
}

func Snap(t T, title, version, content string) {
	t.Helper()
	testName := t.Name()

	// Capture the caller's filename by walking up the call stack
	// to find the first file that's not shutter.go
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

	SnapWithTitle(t, title, testName, fileName, version, content)
}

func SnapWithTitle(t T, title, testName, fileName, version, content string) {
	t.Helper()

	snapshot := &files.Snapshot{
		Title:    title,
		Test:     testName,
		FileName: fileName,
		Content:  content,
		Version:  version,
	}

	accepted, err := files.ReadAccepted(title)
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
