package freeze_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ptdewey/freeze"
)

func TestSnapString(t *testing.T) {
	freeze.SnapString(t, "hello world")
}

func TestSnapMultiple(t *testing.T) {
	freeze.Snap(t, "value1", "value2", 42, "foo", "bar", "baz", "wibble", "wobble", "tick")
}

type CustomStruct struct {
	Name string
	Age  int
}

func (c CustomStruct) Format() string {
	return "CustomStruct{Name: " + c.Name + ", Age: " + string(rune(c.Age)) + "}"
}

func TestSnapCustomType(t *testing.T) {
	cs := CustomStruct{
		Name: "Alice",
		Age:  30,
	}
	freeze.Snap(t, cs)
}

func TestMap(t *testing.T) {
	freeze.Snap(t, map[string]any{
		"foo": "bar",
	})
}

func TestSerializeDeserialize(t *testing.T) {
	snap := &freeze.Snapshot{
		Version:  "1.0.0",
		TestName: "TestExample",
		Content:  "test content\nmultiline",
	}

	serialized := snap.Serialize()
	expected := "---\nversion: 1.0.0\ntest_name: TestExample\n---\ntest content\nmultiline"
	if serialized != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, serialized)
	}

	deserialized, err := freeze.Deserialize(serialized)
	if err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	if deserialized.Version != snap.Version {
		t.Errorf("version mismatch: %s != %s", deserialized.Version, snap.Version)
	}
	if deserialized.TestName != snap.TestName {
		t.Errorf("test name mismatch: %s != %s", deserialized.TestName, snap.TestName)
	}
	if deserialized.Content != snap.Content {
		t.Errorf("content mismatch: %s != %s", deserialized.Content, snap.Content)
	}
}

func TestFileOperations(t *testing.T) {
	snap := &freeze.Snapshot{
		Version:  "0.1.0",
		TestName: "TestFileOps",
		Content:  "file test content",
	}

	if err := freeze.SaveSnapshot(snap, "test"); err != nil {
		t.Fatalf("failed to save snapshot: %v", err)
	}

	read, err := freeze.ReadSnapshot("TestFileOps", "test")
	if err != nil {
		t.Fatalf("failed to read snapshot: %v", err)
	}

	if read.Content != snap.Content {
		t.Errorf("content mismatch: %s != %s", read.Content, snap.Content)
	}

	// cleanupTestSnapshots(t)
}

func TestSnapshotFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"TestMyFunction", "test_my_function"},
		{"test_another_one", "test_another_one"},
		{"TestCamelCase", "test_camel_case"},
		{"TestWithNumbers123", "test_with_numbers123"},
	}

	for _, tt := range tests {
		result := freeze.SnapshotFileName(tt.input)
		if result != tt.expected {
			t.Errorf("SnapshotFileName(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestHistogramDiff(t *testing.T) {
	old := "line1\nline2\nline3"
	new := "line1\nmodified\nline3"

	diff := freeze.Histogram(old, new)

	if len(diff) < 3 {
		t.Errorf("expected at least 3 diff lines, got %d", len(diff))
	}

	if diff[0].Kind != freeze.DiffShared || diff[0].Line != "line1" {
		t.Errorf("line 0: expected shared 'line1', got %v %s", diff[0].Kind, diff[0].Line)
	}

	hasModified := false
	for _, d := range diff {
		if d.Line == "modified" {
			hasModified = true
			if d.Kind != freeze.DiffNew {
				t.Errorf("'modified' should be marked as new")
			}
		}
	}
	if !hasModified {
		t.Error("diff missing 'modified' line")
	}

	hasLine3 := false
	for _, d := range diff {
		if d.Line == "line3" && d.Kind == freeze.DiffShared {
			hasLine3 = true
		}
	}
	if !hasLine3 {
		t.Error("diff should have 'line3' as shared")
	}
}

func TestDiffSnapshotBox(t *testing.T) {
	old := &freeze.Snapshot{
		Version:  "0.1.0",
		TestName: "TestDiff",
		Content:  "old content",
	}

	new := &freeze.Snapshot{
		Version:  "0.1.0",
		TestName: "TestDiff",
		Content:  "new content",
	}

	box := freeze.DiffSnapshotBox(old, new)
	if box == "" {
		t.Error("DiffSnapshotBox returned empty string")
	}

	if !contains(box, "Snapshot Diff") {
		t.Error("DiffSnapshotBox missing header")
	}
}

func TestNewSnapshotBox(t *testing.T) {
	snap := &freeze.Snapshot{
		Version:  "0.1.0",
		TestName: "TestNew",
		Content:  "test content",
	}

	box := freeze.NewSnapshotBox(snap)
	if box == "" {
		t.Error("NewSnapshotBox returned empty string")
	}

	if !contains(box, "New Snapshot") {
		t.Error("NewSnapshotBox missing header")
	}
}

func TestFormatFunctions(t *testing.T) {
	tests := []struct {
		name string
		fn   func(string) string
		text string
	}{
		{"Red", freeze.Red, "error"},
		{"Green", freeze.Green, "success"},
		{"Yellow", freeze.Yellow, "warning"},
		{"Blue", freeze.Blue, "info"},
	}

	for _, tt := range tests {
		result := tt.fn(tt.text)
		if result == "" {
			t.Errorf("%s returned empty string", tt.name)
		}
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func cleanupTestSnapshots(t *testing.T) {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Logf("failed to get cwd: %v", err)
		return
	}

	snapshotDir := filepath.Join(cwd, "__snapshots__")
	_ = os.RemoveAll(snapshotDir)
}
