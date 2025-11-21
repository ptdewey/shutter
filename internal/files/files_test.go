package files_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ptdewey/freeze/internal/files"
)

func TestSnapshotFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"TestMyFunction", "test_my_function"},
		{"test_another_one", "test_another_one"},
		{"TestCamelCase", "test_camel_case"},
		{"TestWithNumbers123", "test_with_numbers123"},
		{"TestABC", "test_a_b_c"},
		{"test", "test"},
		{"TEST", "t_e_s_t"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := files.SnapshotFileName(tt.input)
			if result != tt.expected {
				t.Errorf("SnapshotFileName(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSerializeDeserialize(t *testing.T) {
	snap := &files.Snapshot{
		Title:    "Example Title",
		Test:     "TestExample",
		FileName: "example_test.go",
		Version:  "1.0.0",
		Content:  "test content\nmultiline",
	}

	serialized := snap.Serialize()
	expected := "---\ntitle: Example Title\ntest_name: TestExample\nfile_name: example_test.go\nversion: 1.0.0\n---\ntest content\nmultiline"
	if serialized != expected {
		t.Errorf("Serialize():\nexpected:\n%s\n\ngot:\n%s", expected, serialized)
	}

	deserialized, err := files.Deserialize(serialized)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if deserialized.Title != snap.Title {
		t.Errorf("Title mismatch: %s != %s", deserialized.Title, snap.Title)
	}
	if deserialized.Test != snap.Test {
		t.Errorf("Name mismatch: %s != %s", deserialized.Test, snap.Test)
	}
	if deserialized.FileName != snap.FileName {
		t.Errorf("FileName mismatch: %s != %s", deserialized.FileName, snap.FileName)
	}
	if deserialized.Version != snap.Version {
		t.Errorf("Version mismatch: %s != %s", deserialized.Version, snap.Version)
	}
	if deserialized.Content != snap.Content {
		t.Errorf("Content mismatch: %s != %s", deserialized.Content, snap.Content)
	}
}

func TestDeserializeInvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing separators", "no separators here"},
		{"only one separator", "---\nno closing separator"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := files.Deserialize(tt.input)
			if err == nil {
				t.Error("expected error for invalid format")
			}
		})
	}
}

func TestDeserializeValidFormats(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantTitle   string
		wantTest    string
		wantVersion string
		wantContent string
	}{
		{
			"simple",
			"---\ntitle: Simple Title\ntest_name: Test\nfile_path: /path\nfunc_name: \n---\ncontent",
			"Simple Title",
			"Test",
			"",
			"content",
		},
		{
			"with version",
			"---\ntitle: With Version\ntest_name: Test\nfile_path: /path\nfunc_name: \nversion: 1.0.0\n---\ncontent",
			"With Version",
			"Test",
			"1.0.0",
			"content",
		},
		{
			"multiline content",
			"---\ntitle: Multi Title\ntest_name: MyTest\nfile_path: /path\nfunc_name: \n---\nline1\nline2\nline3",
			"Multi Title",
			"MyTest",
			"",
			"line1\nline2\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snap, err := files.Deserialize(tt.input)
			if err != nil {
				t.Fatalf("Deserialize failed: %v", err)
			}
			if snap.Title != tt.wantTitle {
				t.Errorf("Title = %s, want %s", snap.Title, tt.wantTitle)
			}
			if snap.Test != tt.wantTest {
				t.Errorf("Name = %s, want %s", snap.Test, tt.wantTest)
			}
			if snap.Version != tt.wantVersion {
				t.Errorf("Version = %s, want %s", snap.Version, tt.wantVersion)
			}
			if snap.Content != tt.wantContent {
				t.Errorf("Content = %s, want %s", snap.Content, tt.wantContent)
			}
		})
	}
}

func TestSaveAndReadSnapshot(t *testing.T) {
	snap := &files.Snapshot{
		Title:   "Save Read Title",
		Test:    "TestSaveRead",
		Content: "saved content",
	}

	if err := files.SaveSnapshot(snap, "test"); err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	read, err := files.ReadSnapshot("TestSaveRead", "test")
	if err != nil {
		t.Fatalf("ReadSnapshot failed: %v", err)
	}

	if read.Content != snap.Content {
		t.Errorf("Content mismatch: %s != %s", read.Content, snap.Content)
	}

	cleanupSnapshot(t, "TestSaveRead", "test")
}

func TestReadSnapshotNotFound(t *testing.T) {
	_, err := files.ReadSnapshot("NonExistentTest", "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent snapshot")
	}
}

func TestAcceptSnapshot(t *testing.T) {
	newSnap := &files.Snapshot{
		Title:   "Accept Title",
		Test:    "TestAccept",
		Content: "new content to accept",
	}

	if err := files.SaveSnapshot(newSnap, "new"); err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	if err := files.AcceptSnapshot("TestAccept"); err != nil {
		t.Fatalf("AcceptSnapshot failed: %v", err)
	}

	accepted, err := files.ReadSnapshot("TestAccept", "accepted")
	if err != nil {
		t.Fatalf("ReadSnapshot failed: %v", err)
	}

	if accepted.Content != newSnap.Content {
		t.Errorf("Content mismatch: %s != %s", accepted.Content, newSnap.Content)
	}

	_, err = files.ReadSnapshot("TestAccept", "new")
	if err == nil {
		t.Error("expected error: .new file should be deleted after accept")
	}

	cleanupSnapshot(t, "TestAccept", "accepted")
}

func TestRejectSnapshot(t *testing.T) {
	snap := &files.Snapshot{
		Title:   "Reject Title",
		Test:    "TestReject",
		Content: "content to reject",
	}

	if err := files.SaveSnapshot(snap, "new"); err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	if err := files.RejectSnapshot("TestReject"); err != nil {
		t.Fatalf("RejectSnapshot failed: %v", err)
	}

	_, err := files.ReadSnapshot("TestReject", "new")
	if err == nil {
		t.Error("expected error: .new file should be deleted after reject")
	}
}

func cleanupSnapshot(t *testing.T, testName, state string) {
	t.Helper()

	root, err := os.Getwd()
	if err != nil {
		t.Logf("cleanup: failed to get cwd: %v", err)
		return
	}

	for root != "/" && root != "" {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		root = filepath.Dir(root)
	}

	fileName := files.SnapshotFileName(testName) + "." + state
	filePath := filepath.Join(root, "__snapshots__", fileName)
	_ = os.Remove(filePath)
}
