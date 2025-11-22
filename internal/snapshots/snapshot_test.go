package snapshots

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ptdewey/shutter/internal/files"
)

// mockT is a test implementation that captures test state
type mockT struct {
	helperCalled  bool
	skipCalled    bool
	skipArgs      []any
	skipfCalled   bool
	skipfFormat   string
	skipfArgs     []any
	skipNowCalled bool
	name          string
	errors        []string
	logs          []string
	cleanupFuncs  []func()
}

func (m *mockT) Helper() {
	m.helperCalled = true
}

func (m *mockT) Skip(args ...any) {
	m.skipCalled = true
	m.skipArgs = args
}

func (m *mockT) Skipf(format string, args ...any) {
	m.skipfCalled = true
	m.skipfFormat = format
	m.skipfArgs = args
}

func (m *mockT) SkipNow() {
	m.skipNowCalled = true
}

func (m *mockT) Name() string {
	return m.name
}

func (m *mockT) Error(args ...any) {
	var sb strings.Builder
	for _, arg := range args {
		if s, ok := arg.(string); ok {
			sb.WriteString(s)
		} else {
			sb.WriteString(fmt.Sprintf("%v", arg))
		}
	}
	m.errors = append(m.errors, sb.String())
}

func (m *mockT) Log(args ...any) {
	var sb strings.Builder
	for _, arg := range args {
		if s, ok := arg.(string); ok {
			sb.WriteString(s)
		}
	}
	m.logs = append(m.logs, sb.String())
}

func (m *mockT) Cleanup(f func()) {
	m.cleanupFuncs = append(m.cleanupFuncs, f)
}

func (m *mockT) runCleanups() {
	for _, f := range m.cleanupFuncs {
		f()
	}
}

// Helper to create a temporary snapshot directory for testing
func setupTestDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "shutter-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}

	// Cleanup function
	t.Cleanup(func() {
		os.Chdir(originalDir)
		os.RemoveAll(tmpDir)
	})

	return tmpDir
}

func TestSnap_NewSnapshot(t *testing.T) {
	setupTestDir(t)

	mt := &mockT{name: "TestExample"}
	Snap(mt, "test_snap", "v1", "content here")

	// Should create a new snapshot
	if len(mt.errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(mt.errors))
	}

	if !strings.Contains(mt.errors[0], "new snapshot created") {
		t.Errorf("expected 'new snapshot created' error, got: %s", mt.errors[0])
	}

	// Verify snapshot file was created
	snapPath := filepath.Join("__snapshots__", "test_snap.snap.new")
	if _, err := os.Stat(snapPath); os.IsNotExist(err) {
		t.Error("expected snapshot file to be created")
	}
}

func TestSnap_MatchingSnapshot(t *testing.T) {
	setupTestDir(t)

	// Create an accepted snapshot first
	accepted := &files.Snapshot{
		Title:    "matching_test",
		Test:     "TestExample",
		FileName: "test.go",
		Content:  "expected content",
		Version:  "v1",
	}
	if err := files.SaveSnapshot(accepted, "accepted"); err != nil {
		t.Fatalf("failed to save accepted snapshot: %v", err)
	}

	mt := &mockT{name: "TestExample"}
	Snap(mt, "matching_test", "v1", "expected content")

	// Should not report any errors (snapshot matches)
	if len(mt.errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(mt.errors), mt.errors)
	}
}

func TestSnap_MismatchedSnapshot(t *testing.T) {
	setupTestDir(t)

	// Create an accepted snapshot
	accepted := &files.Snapshot{
		Title:    "mismatched_test",
		Test:     "TestExample",
		FileName: "test.go",
		Content:  "old content",
		Version:  "v1",
	}
	if err := files.SaveSnapshot(accepted, "accepted"); err != nil {
		t.Fatalf("failed to save accepted snapshot: %v", err)
	}

	mt := &mockT{name: "TestExample"}
	Snap(mt, "mismatched_test", "v1", "new content")

	// Should report a mismatch error
	if len(mt.errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(mt.errors))
	}

	if !strings.Contains(mt.errors[0], "snapshot mismatch") {
		t.Errorf("expected 'snapshot mismatch' error, got: %s", mt.errors[0])
	}

	// Verify new snapshot file was created
	snapPath := filepath.Join("__snapshots__", "mismatched_test.snap.new")
	if _, err := os.Stat(snapPath); os.IsNotExist(err) {
		t.Error("expected new snapshot file to be created")
	}
}

func TestSnap_CallerDetection(t *testing.T) {
	setupTestDir(t)

	mt := &mockT{name: "TestCallerDetection"}
	Snap(mt, "caller_test", "v1", "test content")

	// Read the created snapshot
	snap, err := files.ReadSnapshot("caller_test", "new")
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}

	// Should detect the caller filename (this test file)
	if snap.FileName == "unknown" {
		t.Error("expected caller filename to be detected")
	}

	// Should not be shutter.go
	if snap.FileName == "shutter.go" {
		t.Error("caller should not be shutter.go")
	}
}

func TestSnapWithTitle_CreatesCorrectSnapshot(t *testing.T) {
	setupTestDir(t)

	mt := &mockT{name: "TestExample"}
	SnapWithTitle(mt, "custom_title", "TestExample", "test.go", "v1", "custom content")

	// Read the snapshot
	snap, err := files.ReadSnapshot("custom_title", "new")
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}

	if snap.Title != "custom_title" {
		t.Errorf("expected title 'custom_title', got %q", snap.Title)
	}
	if snap.Test != "TestExample" {
		t.Errorf("expected test name 'TestExample', got %q", snap.Test)
	}
	if snap.FileName != "test.go" {
		t.Errorf("expected filename 'test.go', got %q", snap.FileName)
	}
	if snap.Version != "v1" {
		t.Errorf("expected version 'v1', got %q", snap.Version)
	}
	if snap.Content != "custom content" {
		t.Errorf("expected content 'custom content', got %q", snap.Content)
	}
}

func TestSnapWithTitle_MatchingContent(t *testing.T) {
	setupTestDir(t)

	// Create accepted snapshot
	accepted := &files.Snapshot{
		Title:    "match_title",
		Test:     "TestMatch",
		FileName: "test.go",
		Content:  "same content",
		Version:  "v1",
	}
	if err := files.SaveSnapshot(accepted, "accepted"); err != nil {
		t.Fatalf("failed to save accepted snapshot: %v", err)
	}

	mt := &mockT{name: "TestMatch"}
	SnapWithTitle(mt, "match_title", "TestMatch", "test.go", "v1", "same content")

	// Should not error (content matches)
	if len(mt.errors) != 0 {
		t.Errorf("expected no errors for matching content, got: %v", mt.errors)
	}
}

func TestSnapWithTitle_MismatchedContent(t *testing.T) {
	setupTestDir(t)

	// Create accepted snapshot
	accepted := &files.Snapshot{
		Title:    "mismatch_title",
		Test:     "TestMismatch",
		FileName: "test.go",
		Content:  "old content",
		Version:  "v1",
	}
	if err := files.SaveSnapshot(accepted, "accepted"); err != nil {
		t.Fatalf("failed to save accepted snapshot: %v", err)
	}

	mt := &mockT{name: "TestMismatch"}
	SnapWithTitle(mt, "mismatch_title", "TestMismatch", "test.go", "v1", "new content")

	// Should error about mismatch
	if len(mt.errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(mt.errors))
	}
	if !strings.Contains(mt.errors[0], "snapshot mismatch") {
		t.Errorf("expected mismatch error, got: %s", mt.errors[0])
	}

	// Should create new snapshot file
	newSnap, err := files.ReadSnapshot("mismatch_title", "new")
	if err != nil {
		t.Fatalf("failed to read new snapshot: %v", err)
	}
	if newSnap.Content != "new content" {
		t.Errorf("expected new content in .snap.new file")
	}
}

func TestSnap_HelperCalled(t *testing.T) {
	setupTestDir(t)

	mt := &mockT{name: "TestHelper"}
	Snap(mt, "helper_test", "v1", "content")

	if !mt.helperCalled {
		t.Error("expected Helper() to be called")
	}
}

func TestSnapWithTitle_HelperCalled(t *testing.T) {
	setupTestDir(t)

	mt := &mockT{name: "TestHelper"}
	SnapWithTitle(mt, "helper_test", "TestHelper", "test.go", "v1", "content")

	if !mt.helperCalled {
		t.Error("expected Helper() to be called")
	}
}

func TestSnap_MultipleDifferentSnapshots(t *testing.T) {
	setupTestDir(t)

	mt := &mockT{name: "TestMultiple"}

	// Create multiple snapshots
	Snap(mt, "snap_one", "v1", "content one")
	Snap(mt, "snap_two", "v1", "content two")
	Snap(mt, "snap_three", "v1", "content three")

	// All should be created as new snapshots
	if len(mt.errors) != 3 {
		t.Errorf("expected 3 errors (new snapshots), got %d", len(mt.errors))
	}

	// Verify all files exist
	for _, title := range []string{"snap_one", "snap_two", "snap_three"} {
		snapPath := filepath.Join("__snapshots__", title+".snap.new")
		if _, err := os.Stat(snapPath); os.IsNotExist(err) {
			t.Errorf("expected snapshot %s to exist", title)
		}
	}
}

func TestSnap_EmptyContent(t *testing.T) {
	setupTestDir(t)

	mt := &mockT{name: "TestEmpty"}
	Snap(mt, "empty_test", "v1", "")

	// Should create snapshot with empty content
	snap, err := files.ReadSnapshot("empty_test", "new")
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}
	if snap.Content != "" {
		t.Errorf("expected empty content, got %q", snap.Content)
	}
}

func TestSnap_MultilineContent(t *testing.T) {
	setupTestDir(t)

	content := `line one
line two
line three`

	mt := &mockT{name: "TestMultiline"}
	Snap(mt, "multiline_test", "v1", content)

	snap, err := files.ReadSnapshot("multiline_test", "new")
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}
	if snap.Content != content {
		t.Errorf("expected multiline content to match")
	}
}

func TestSnap_SpecialCharacters(t *testing.T) {
	setupTestDir(t)

	content := `{"key": "value with \"quotes\"", "emoji": "🎉"}`

	mt := &mockT{name: "TestSpecial"}
	Snap(mt, "special_test", "v1", content)

	snap, err := files.ReadSnapshot("special_test", "new")
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}
	if snap.Content != content {
		t.Errorf("expected special characters to be preserved")
	}
}

func TestSnap_VersionTracking(t *testing.T) {
	setupTestDir(t)

	mt := &mockT{name: "TestVersion"}
	Snap(mt, "version_test", "v2", "content")

	snap, err := files.ReadSnapshot("version_test", "new")
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}
	if snap.Version != "v2" {
		t.Errorf("expected version 'v2', got %q", snap.Version)
	}
}

func TestSnap_UpdateAcceptedSnapshot(t *testing.T) {
	setupTestDir(t)

	// Create initial accepted snapshot
	accepted := &files.Snapshot{
		Title:    "update_test",
		Test:     "TestUpdate",
		FileName: "test.go",
		Content:  "old version",
		Version:  "v1",
	}
	if err := files.SaveSnapshot(accepted, "accepted"); err != nil {
		t.Fatalf("failed to save accepted snapshot: %v", err)
	}

	// Create new snapshot with different content
	mt := &mockT{name: "TestUpdate"}
	Snap(mt, "update_test", "v2", "new version")

	// Verify new snapshot was created
	newSnap, err := files.ReadSnapshot("update_test", "new")
	if err != nil {
		t.Fatalf("failed to read new snapshot: %v", err)
	}

	if newSnap.Content != "new version" {
		t.Errorf("expected new content")
	}
	if newSnap.Version != "v2" {
		t.Errorf("expected version to be updated")
	}

	// Accepted snapshot should remain unchanged
	acceptedSnap, err := files.ReadSnapshot("update_test", "snap")
	if err != nil {
		t.Fatalf("failed to read accepted snapshot: %v", err)
	}
	if acceptedSnap.Content != "old version" {
		t.Errorf("accepted snapshot should not change")
	}
}

func TestSnap_WithSpacesInTitle(t *testing.T) {
	setupTestDir(t)

	mt := &mockT{name: "TestSpaces"}
	Snap(mt, "test with spaces", "v1", "content")

	// Should normalize title to filename
	snap, err := files.ReadSnapshot("test with spaces", "new")
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}

	if snap.Title != "test with spaces" {
		t.Errorf("expected title to preserve spaces, got %q", snap.Title)
	}
}
