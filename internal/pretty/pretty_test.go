package pretty_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ptdewey/shutter/internal/diff"
	"github.com/ptdewey/shutter/internal/files"
	"github.com/ptdewey/shutter/internal/pretty"
)

func TestColorFunctionsWithColor(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	tests := []struct {
		name string
		fn   func(string) string
		text string
	}{
		{"Red", pretty.Red, "error"},
		{"Green", pretty.Green, "success"},
		{"Yellow", pretty.Yellow, "warning"},
		{"Blue", pretty.Blue, "info"},
		{"Gray", pretty.Gray, "gray"},
		{"Bold", pretty.Bold, "bold"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.text)
			if result == "" {
				t.Errorf("%s returned empty string", tt.name)
			}
			if result == tt.text {
				t.Errorf("%s did not add color codes", tt.name)
			}
			if !contains(result, tt.text) {
				t.Errorf("%s does not contain original text", tt.name)
			}
		})
	}
}

func TestColorFunctionsNoColor(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	tests := []struct {
		name string
		fn   func(string) string
		text string
	}{
		{"Red", pretty.Red, "error"},
		{"Green", pretty.Green, "success"},
		{"Yellow", pretty.Yellow, "warning"},
		{"Blue", pretty.Blue, "info"},
		{"Gray", pretty.Gray, "gray"},
		{"Bold", pretty.Bold, "bold"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.text)
			if result != tt.text {
				t.Errorf("%s should return plain text when NO_COLOR is set", tt.name)
			}
		})
	}
}

func TestHeader(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := pretty.Header("test header")
	if result == "" {
		t.Error("Header returned empty string")
	}
	if result == "test header" {
		t.Error("Header should apply formatting")
	}
	if !contains(result, "test header") {
		t.Error("Header should contain original text")
	}
}

func TestSuccess(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := pretty.Success("success message")
	if result == "" {
		t.Error("Success returned empty string")
	}
	if !contains(result, "success message") {
		t.Error("Success should contain original text")
	}
}

func TestError(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := pretty.Error("error message")
	if result == "" {
		t.Error("Error returned empty string")
	}
	if !contains(result, "error message") {
		t.Error("Error should contain original text")
	}
}

func TestWarning(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	result := pretty.Warning("warning message")
	if result == "" {
		t.Error("Warning returned empty string")
	}
	if !contains(result, "warning message") {
		t.Error("Warning should contain original text")
	}
}

func TestTerminalWidth(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected int
	}{
		{"default", "", 80},
		{"valid width", "120", 120},
		{"invalid width", "invalid", 80},
		{"zero width", "0", 80},
		{"negative width", "-10", 80},
		{"large width", "1000", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue == "" {
				os.Unsetenv("COLUMNS")
			} else {
				os.Setenv("COLUMNS", tt.envValue)
			}
			defer os.Unsetenv("COLUMNS")

			result := pretty.TerminalWidth()
			if result != tt.expected {
				t.Errorf("TerminalWidth() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDiffSnapshotBox(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "100")
	defer os.Unsetenv("COLUMNS")

	oldContent := `line1
line2
line3`

	newContent := `line1
modified
line3`

	oldSnap := &files.Snapshot{
		Title:   "Test Snapshot",
		Test:    "TestExample",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "Test Snapshot",
		Test:    "TestExample",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	// Check that result is not empty
	if result == "" {
		t.Error("DiffSnapshotBox returned empty string")
	}

	// Check for header elements
	if !strings.Contains(result, "Snapshot Diff") {
		t.Error("Result should contain 'Snapshot Diff' header")
	}
	if !strings.Contains(result, "Test Snapshot") {
		t.Error("Result should contain title")
	}
	if !strings.Contains(result, "TestExample") {
		t.Error("Result should contain test name")
	}

	// Check for diff content
	if !strings.Contains(result, "line1") {
		t.Error("Result should contain 'line1'")
	}
	if !strings.Contains(result, "modified") {
		t.Error("Result should contain 'modified'")
	}
	if !strings.Contains(result, "line3") {
		t.Error("Result should contain 'line3'")
	}

	// Print the result for visual inspection
	t.Logf("\n%s", result)
}

func TestDiffSnapshotBoxLargeLineNumbers(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "120")
	defer os.Unsetenv("COLUMNS")

	// Create content with more than 10 lines to test multi-digit line numbers
	oldLines := make([]string, 15)
	newLines := make([]string, 15)
	for i := 0; i < 15; i++ {
		oldLines[i] = fmt.Sprintf("line %d", i+1)
		newLines[i] = fmt.Sprintf("line %d", i+1)
	}
	// Modify line 10
	oldLines[9] = "line 10 old"
	newLines[9] = "line 10 new"

	oldContent := strings.Join(oldLines, "\n")
	newContent := strings.Join(newLines, "\n")

	oldSnap := &files.Snapshot{
		Title:   "Large Diff Test",
		Test:    "TestLargeDiff",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "Large Diff Test",
		Test:    "TestLargeDiff",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	// Check that result is not empty
	if result == "" {
		t.Error("DiffSnapshotBox returned empty string")
	}

	// Print the result for visual inspection
	t.Logf("\n%s", result)
}
