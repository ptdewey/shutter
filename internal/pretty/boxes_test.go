package pretty_test

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/ptdewey/shutter"
	"github.com/ptdewey/shutter/internal/diff"
	"github.com/ptdewey/shutter/internal/files"
	"github.com/ptdewey/shutter/internal/pretty"
)

// BoxValidation holds expected properties for validation
type BoxValidation struct {
	// Title and filename expectations
	Title       string
	TestName    string
	FileName    string
	HasTitle    bool
	HasTestName bool
	HasFileName bool

	// Diff line expectations
	ExpectedAdds    []string // Lines that should appear as additions (green +)
	ExpectedDeletes []string // Lines that should appear as deletions (red -)
	ExpectedContext []string // Lines that should appear as context (gray │)

	// Structural expectations
	HasTopBar    bool
	HasBottomBar bool
	MinLines     int // Minimum number of content lines expected
}

// ValidateDiffBox checks that a diff box output matches expectations
func ValidateDiffBox(t *testing.T, output string, validation BoxValidation) {
	t.Helper()

	// Remove ANSI codes for easier content checking
	stripped := stripANSI(output)

	// Check title/test/filename presence
	if validation.HasTitle {
		if !strings.Contains(stripped, "title: "+validation.Title) {
			t.Errorf("Expected title '%s' not found in output", validation.Title)
		}
	}

	if validation.HasTestName {
		if !strings.Contains(stripped, "test: "+validation.TestName) {
			t.Errorf("Expected test name '%s' not found in output", validation.TestName)
		}
	}

	if validation.HasFileName {
		if !strings.Contains(stripped, "file: "+validation.FileName) {
			t.Errorf("Expected file name '%s' not found in output", validation.FileName)
		}
	}

	// Check for box structure
	if validation.HasTopBar {
		if !strings.Contains(stripped, "┬") {
			t.Error("Expected top bar with ┬ character")
		}
	}

	if validation.HasBottomBar {
		if !strings.Contains(stripped, "┴") {
			t.Error("Expected bottom bar with ┴ character")
		}
	}

	// Check expected additions (green + lines)
	for _, expectedAdd := range validation.ExpectedAdds {
		if !containsDiffLine(output, "+", expectedAdd) {
			t.Errorf("Expected addition not found: + %s", expectedAdd)
		}
	}

	// Check expected deletions (red - lines)
	for _, expectedDelete := range validation.ExpectedDeletes {
		if !containsDiffLine(output, "-", expectedDelete) {
			t.Errorf("Expected deletion not found: - %s", expectedDelete)
		}
	}

	// Check expected context (shared lines)
	for _, expectedContext := range validation.ExpectedContext {
		if !containsDiffLine(output, "│", expectedContext) {
			t.Errorf("Expected context line not found: │ %s", expectedContext)
		}
	}

	// Check minimum line count
	if validation.MinLines > 0 {
		lines := strings.Split(output, "\n")
		contentLines := countContentLines(lines)
		if contentLines < validation.MinLines {
			t.Errorf("Expected at least %d content lines, got %d", validation.MinLines, contentLines)
		}
	}
}

// containsDiffLine checks if a line with the given prefix and content exists
func containsDiffLine(output, prefix, content string) bool {
	lines := strings.Split(output, "\n")
	stripped := stripANSI(output)
	strippedLines := strings.Split(stripped, "\n")

	for i, line := range strippedLines {
		// Check if line contains the prefix and content
		if strings.Contains(line, prefix) && strings.Contains(line, content) {
			// Verify the original line has proper coloring
			originalLine := lines[i]
			switch prefix {
			case "+":
				// Green additions should have ANSI codes
				if !strings.Contains(originalLine, "\033[") {
					continue // Skip if no color
				}
			case "-":
				// Red deletions should have ANSI codes
				if !strings.Contains(originalLine, "\033[") {
					continue
				}
			case "│":
				// Context lines may or may not have color
			}
			return true
		}
	}
	return false
}

// countContentLines counts lines that contain diff content (not headers/borders)
func countContentLines(lines []string) int {
	count := 0
	for _, line := range lines {
		stripped := stripANSI(line)
		// Content lines have line numbers followed by +, -, or │
		if strings.Contains(stripped, "+") ||
			strings.Contains(stripped, "-") ||
			strings.Contains(stripped, "│") {
			count++
		}
	}
	return count
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

// TestDiffSnapshotBox_SimpleModification tests a basic modification scenario
func TestDiffSnapshotBox_SimpleModification(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "100")
	defer os.Unsetenv("COLUMNS")

	oldContent := "line1\nline2\nline3"
	newContent := "line1\nmodified\nline3"

	oldSnap := &files.Snapshot{
		Title:   "Simple Modification",
		Test:    "TestSimple",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "Simple Modification",
		Test:    "TestSimple",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	validation := BoxValidation{
		Title:           "Simple Modification",
		TestName:        "TestSimple",
		FileName:        "testsimple.snap",
		HasTitle:        true,
		HasTestName:     true,
		HasFileName:     true,
		ExpectedAdds:    []string{"modified"},
		ExpectedDeletes: []string{"line2"},
		ExpectedContext: []string{"line1", "line3"},
		HasTopBar:       true,
		HasBottomBar:    true,
		MinLines:        4, // 1 shared + 1 delete + 1 add + 1 shared
	}

	ValidateDiffBox(t, result, validation)
}

// TestDiffSnapshotBox_PureAddition tests adding lines only
func TestDiffSnapshotBox_PureAddition(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "100")
	defer os.Unsetenv("COLUMNS")

	oldContent := "line1\nline2"
	newContent := "line1\nline2\nline3\nline4"

	oldSnap := &files.Snapshot{
		Title:   "Pure Addition",
		Test:    "TestAddition",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "Pure Addition",
		Test:    "TestAddition",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	validation := BoxValidation{
		Title:           "Pure Addition",
		TestName:        "TestAddition",
		FileName:        "testaddition.snap",
		HasTitle:        true,
		HasTestName:     true,
		HasFileName:     true,
		ExpectedAdds:    []string{"line3", "line4"},
		ExpectedDeletes: []string{},
		ExpectedContext: []string{"line1", "line2"},
		HasTopBar:       true,
		HasBottomBar:    true,
		MinLines:        4, // 2 shared + 2 adds
	}

	ValidateDiffBox(t, result, validation)
}

// TestDiffSnapshotBox_PureDeletion tests deleting lines only
func TestDiffSnapshotBox_PureDeletion(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "100")
	defer os.Unsetenv("COLUMNS")

	oldContent := "line1\nline2\nline3\nline4"
	newContent := "line1\nline2"

	oldSnap := &files.Snapshot{
		Title:   "Pure Deletion",
		Test:    "TestDeletion",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "Pure Deletion",
		Test:    "TestDeletion",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	validation := BoxValidation{
		Title:           "Pure Deletion",
		TestName:        "TestDeletion",
		FileName:        "testdeletion.snap",
		HasTitle:        true,
		HasTestName:     true,
		HasFileName:     true,
		ExpectedAdds:    []string{},
		ExpectedDeletes: []string{"line3", "line4"},
		ExpectedContext: []string{"line1", "line2"},
		HasTopBar:       true,
		HasBottomBar:    true,
		MinLines:        4, // 2 shared + 2 deletes
	}

	ValidateDiffBox(t, result, validation)
}

// TestDiffSnapshotBox_ComplexMixed tests multiple types of changes
func TestDiffSnapshotBox_ComplexMixed(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "120")
	defer os.Unsetenv("COLUMNS")

	oldContent := `unchanged1
delete1
delete2
unchanged2
modify_old
unchanged3`

	newContent := `unchanged1
unchanged2
modify_new
add1
unchanged3
add2`

	oldSnap := &files.Snapshot{
		Title:   "Complex Mixed",
		Test:    "TestComplexMixed",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "Complex Mixed",
		Test:    "TestComplexMixed",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	validation := BoxValidation{
		Title:           "Complex Mixed",
		TestName:        "TestComplexMixed",
		FileName:        "testcomplexmixed.snap",
		HasTitle:        true,
		HasTestName:     true,
		HasFileName:     true,
		ExpectedAdds:    []string{"modify_new", "add1", "add2"},
		ExpectedDeletes: []string{"delete1", "delete2", "modify_old"},
		ExpectedContext: []string{"unchanged1", "unchanged2", "unchanged3"},
		HasTopBar:       true,
		HasBottomBar:    true,
		MinLines:        9, // 3 shared + 3 deletes + 3 adds
	}

	ValidateDiffBox(t, result, validation)
}

// TestDiffSnapshotBox_EmptyOld tests diff from empty to content
func TestDiffSnapshotBox_EmptyOld(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "100")
	defer os.Unsetenv("COLUMNS")

	oldContent := ""
	newContent := "line1\nline2\nline3"

	oldSnap := &files.Snapshot{
		Title:   "Empty to Content",
		Test:    "TestEmptyOld",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "Empty to Content",
		Test:    "TestEmptyOld",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	validation := BoxValidation{
		Title:           "Empty to Content",
		TestName:        "TestEmptyOld",
		FileName:        "testemptyold.snap",
		HasTitle:        true,
		HasTestName:     true,
		HasFileName:     true,
		ExpectedAdds:    []string{"line1", "line2", "line3"},
		ExpectedDeletes: []string{},
		ExpectedContext: []string{},
		HasTopBar:       true,
		HasBottomBar:    true,
		MinLines:        3, // 3 adds
	}

	ValidateDiffBox(t, result, validation)
}

// TestDiffSnapshotBox_EmptyNew tests diff from content to empty
func TestDiffSnapshotBox_EmptyNew(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "100")
	defer os.Unsetenv("COLUMNS")

	oldContent := "line1\nline2\nline3"
	newContent := ""

	oldSnap := &files.Snapshot{
		Title:   "Content to Empty",
		Test:    "TestEmptyNew",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "Content to Empty",
		Test:    "TestEmptyNew",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	validation := BoxValidation{
		Title:           "Content to Empty",
		TestName:        "TestEmptyNew",
		FileName:        "testemptynew.snap",
		HasTitle:        true,
		HasTestName:     true,
		HasFileName:     true,
		ExpectedAdds:    []string{},
		ExpectedDeletes: []string{"line1", "line2", "line3"},
		ExpectedContext: []string{},
		HasTopBar:       true,
		HasBottomBar:    true,
		MinLines:        3, // 3 deletes
	}

	ValidateDiffBox(t, result, validation)
}

// TestDiffSnapshotBox_NoTitle tests snapshot without title
func TestDiffSnapshotBox_NoTitle(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "100")
	defer os.Unsetenv("COLUMNS")

	oldContent := "old"
	newContent := "new"

	oldSnap := &files.Snapshot{
		Title:   "",
		Test:    "TestNoTitle",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "",
		Test:    "TestNoTitle",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	stripped := stripANSI(result)

	// Should NOT contain "title:" line
	if strings.Contains(stripped, "title:") {
		t.Error("Expected no title line when title is empty")
	}

	// Should still contain test and file
	if !strings.Contains(stripped, "test: TestNoTitle") {
		t.Error("Expected test name to be present")
	}
	if !strings.Contains(stripped, "file: testnotitle.snap") {
		t.Error("Expected file name to be present")
	}
}

// TestDiffSnapshotBox_LargeLineNumbers tests proper padding for multi-digit line numbers
func TestDiffSnapshotBox_LargeLineNumbers(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "120")
	defer os.Unsetenv("COLUMNS")

	// Create content with 100+ lines to test 3-digit line numbers
	oldLines := make([]string, 105)
	newLines := make([]string, 105)
	for i := 0; i < 105; i++ {
		oldLines[i] = fmt.Sprintf("line %d", i+1)
		newLines[i] = fmt.Sprintf("line %d", i+1)
	}
	// Modify lines 50, 75, and 100
	oldLines[49] = "old line 50"
	newLines[49] = "new line 50"
	oldLines[74] = "old line 75"
	newLines[74] = "new line 75"
	oldLines[99] = "old line 100"
	newLines[99] = "new line 100"

	oldContent := strings.Join(oldLines, "\n")
	newContent := strings.Join(newLines, "\n")

	oldSnap := &files.Snapshot{
		Title:   "Large Line Numbers",
		Test:    "TestLargeLineNumbers",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "Large Line Numbers",
		Test:    "TestLargeLineNumbers",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	stripped := stripANSI(result)

	// Check that 3-digit line numbers appear
	if !strings.Contains(stripped, "100") {
		t.Error("Expected 3-digit line number 100 to appear")
	}

	// Validate line number alignment by checking that numbers are right-aligned
	// Line 1 should have padding for 3 digits
	lines := strings.Split(stripped, "\n")
	foundSingleDigit := false
	foundTripleDigit := false

	for _, line := range lines {
		// Look for lines with content markers
		if strings.Contains(line, "│") || strings.Contains(line, "+") || strings.Contains(line, "-") {
			// Single digit should have padding (e.g., "  1" or "  2")
			if strings.Contains(line, "  1 ") || strings.Contains(line, "  2 ") {
				foundSingleDigit = true
			}
			// Triple digit should align (e.g., "100" or "105")
			if strings.Contains(line, "100 ") || strings.Contains(line, "105 ") {
				foundTripleDigit = true
			}
		}
	}

	if !foundSingleDigit {
		t.Error("Expected to find padded single-digit line numbers")
	}
	if !foundTripleDigit {
		t.Error("Expected to find triple-digit line numbers")
	}
}

// TestDiffSnapshotBox_UnicodeContent tests diff with unicode characters
func TestDiffSnapshotBox_UnicodeContent(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "100")
	defer os.Unsetenv("COLUMNS")

	oldContent := "Hello 世界\nこんにちは\n🎉 emoji"
	newContent := "Hello 世界\nさようなら\n🎊 party"

	oldSnap := &files.Snapshot{
		Title:   "Unicode Test",
		Test:    "TestUnicode",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "Unicode Test",
		Test:    "TestUnicode",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	validation := BoxValidation{
		Title:           "Unicode Test",
		TestName:        "TestUnicode",
		FileName:        "testunicode.snap",
		HasTitle:        true,
		HasTestName:     true,
		HasFileName:     true,
		ExpectedAdds:    []string{"さようなら", "🎊 party"},
		ExpectedDeletes: []string{"こんにちは", "🎉 emoji"},
		ExpectedContext: []string{"Hello 世界"},
		HasTopBar:       true,
		HasBottomBar:    true,
		MinLines:        5, // 1 shared + 2 deletes + 2 adds
	}

	ValidateDiffBox(t, result, validation)
}

// TestNewSnapshotBox_Basic tests the new snapshot box rendering
func TestNewSnapshotBox_Basic(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "100")
	defer os.Unsetenv("COLUMNS")

	content := "line1\nline2\nline3"

	snap := &files.Snapshot{
		Title:    "New Snapshot",
		Test:     "TestNewSnapshot",
		FileName: "test_new.snap",
		Content:  content,
	}

	result := pretty.NewSnapshotBox(snap)

	stripped := stripANSI(result)

	// Check header
	if !strings.Contains(stripped, "New Snapshot") {
		t.Error("Expected 'New Snapshot' header")
	}

	// Check metadata
	if !strings.Contains(stripped, "title: New Snapshot") {
		t.Error("Expected title in output")
	}
	if !strings.Contains(stripped, "test: TestNewSnapshot") {
		t.Error("Expected test name in output")
	}
	if !strings.Contains(stripped, "file: test_new.snap") {
		t.Error("Expected file name in output")
	}

	// Check content lines (all should be green additions)
	if !containsDiffLine(result, "+", "line1") {
		t.Error("Expected line1 as addition")
	}
	if !containsDiffLine(result, "+", "line2") {
		t.Error("Expected line2 as addition")
	}
	if !containsDiffLine(result, "+", "line3") {
		t.Error("Expected line3 as addition")
	}

	// Check box structure
	if !strings.Contains(stripped, "┬") {
		t.Error("Expected top bar with ┬")
	}
	if !strings.Contains(stripped, "┴") {
		t.Error("Expected bottom bar with ┴")
	}
}

// TestNewSnapshotBox_EmptyContent tests new snapshot with empty content
func TestNewSnapshotBox_EmptyContent(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "100")
	defer os.Unsetenv("COLUMNS")

	snap := &files.Snapshot{
		Title:    "Empty Snapshot",
		Test:     "TestEmpty",
		FileName: "test_empty.snap",
		Content:  "",
	}

	result := pretty.NewSnapshotBox(snap)

	// Should still render box with metadata, just no content lines
	stripped := stripANSI(result)

	if !strings.Contains(stripped, "title: Empty Snapshot") {
		t.Error("Expected title in output")
	}

	// Should have box structure even with empty content
	if !strings.Contains(stripped, "┬") {
		t.Error("Expected top bar with ┬")
	}
	if !strings.Contains(stripped, "┴") {
		t.Error("Expected bottom bar with ┴")
	}
}

// Snapshot testing for visual regression

func TestDiffSnapshotBox_VisualRegression_SimpleModification(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "100")
	defer os.Unsetenv("COLUMNS")

	oldContent := "line1\nline2\nline3"
	newContent := "line1\nmodified\nline3"

	oldSnap := &files.Snapshot{
		Title:   "Visual Test",
		Test:    "TestVisualSimple",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "Visual Test",
		Test:    "TestVisualSimple",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	shutter.SnapString(t, "diff_box_simple_modification", result)
}

func TestDiffSnapshotBox_VisualRegression_ComplexMixed(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "120")
	defer os.Unsetenv("COLUMNS")

	oldContent := `unchanged1
delete1
delete2
unchanged2
modify_old
unchanged3`

	newContent := `unchanged1
unchanged2
modify_new
add1
unchanged3
add2`

	oldSnap := &files.Snapshot{
		Title:   "Visual Complex",
		Test:    "TestVisualComplex",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "Visual Complex",
		Test:    "TestVisualComplex",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	shutter.SnapString(t, "diff_box_complex_mixed", result)
}

func TestDiffSnapshotBox_VisualRegression_LargeLineNumbers(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "120")
	defer os.Unsetenv("COLUMNS")

	// Create content with 100+ lines
	oldLines := make([]string, 105)
	newLines := make([]string, 105)
	for i := 0; i < 105; i++ {
		oldLines[i] = fmt.Sprintf("line %d", i+1)
		newLines[i] = fmt.Sprintf("line %d", i+1)
	}
	oldLines[49] = "old line 50"
	newLines[49] = "new line 50"
	oldLines[99] = "old line 100"
	newLines[99] = "new line 100"

	oldContent := strings.Join(oldLines, "\n")
	newContent := strings.Join(newLines, "\n")

	oldSnap := &files.Snapshot{
		Title:   "Large Line Numbers",
		Test:    "TestVisualLarge",
		Content: oldContent,
	}

	newSnap := &files.Snapshot{
		Title:   "Large Line Numbers",
		Test:    "TestVisualLarge",
		Content: newContent,
	}

	diffLines := diff.Histogram(oldContent, newContent)
	result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

	shutter.SnapString(t, "diff_box_large_line_numbers", result)
}

func TestNewSnapshotBox_VisualRegression(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "100")
	defer os.Unsetenv("COLUMNS")

	content := "line1\nline2\nline3\nline4\nline5"

	snap := &files.Snapshot{
		Title:    "New Snapshot Visual",
		Test:     "TestNewVisual",
		FileName: "test_new_visual.snap",
		Content:  content,
	}

	result := pretty.NewSnapshotBox(snap)

	shutter.SnapString(t, "new_snapshot_box", result)
}

// Randomized testing

func TestDiffSnapshotBox_Random_Additions(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "120")
	defer os.Unsetenv("COLUMNS")

	rng := rand.New(rand.NewSource(12345)) // Fixed seed for reproducibility

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("random_addition_%d", i), func(t *testing.T) {
			// Generate random number of old lines (5-20)
			numOldLines := rng.Intn(16) + 5
			oldLines := make([]string, numOldLines)
			for j := 0; j < numOldLines; j++ {
				oldLines[j] = fmt.Sprintf("old_line_%d", j+1)
			}

			// Add random number of new lines (1-10)
			numNewLines := rng.Intn(10) + 1
			newLines := make([]string, numOldLines+numNewLines)
			copy(newLines, oldLines)
			for j := 0; j < numNewLines; j++ {
				newLines[numOldLines+j] = fmt.Sprintf("new_line_%d", j+1)
			}

			oldContent := strings.Join(oldLines, "\n")
			newContent := strings.Join(newLines, "\n")

			oldSnap := &files.Snapshot{
				Title:   fmt.Sprintf("Random Addition %d", i),
				Test:    fmt.Sprintf("TestRandomAdd_%d", i),
				Content: oldContent,
			}

			newSnap := &files.Snapshot{
				Title:   fmt.Sprintf("Random Addition %d", i),
				Test:    fmt.Sprintf("TestRandomAdd_%d", i),
				Content: newContent,
			}

			diffLines := diff.Histogram(oldContent, newContent)
			result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

			// Validate structure
			stripped := stripANSI(result)

			// Should have box structure
			if !strings.Contains(stripped, "┬") {
				t.Error("Missing top bar")
			}
			if !strings.Contains(stripped, "┴") {
				t.Error("Missing bottom bar")
			}

			// Should contain title and test name
			if !strings.Contains(stripped, fmt.Sprintf("Random Addition %d", i)) {
				t.Error("Missing title")
			}

			// Count additions
			addCount := strings.Count(result, "+")
			if addCount < numNewLines {
				t.Errorf("Expected at least %d additions, got %d", numNewLines, addCount)
			}
		})
	}
}

func TestDiffSnapshotBox_Random_Deletions(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "120")
	defer os.Unsetenv("COLUMNS")

	rng := rand.New(rand.NewSource(54321))

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("random_deletion_%d", i), func(t *testing.T) {
			// Generate random number of old lines (10-30)
			numOldLines := rng.Intn(21) + 10
			oldLines := make([]string, numOldLines)
			for j := 0; j < numOldLines; j++ {
				oldLines[j] = fmt.Sprintf("line_%d", j+1)
			}

			// Delete random number of lines (1-5)
			numToDelete := rng.Intn(5) + 1
			if numToDelete > numOldLines {
				numToDelete = numOldLines / 2
			}
			newLines := make([]string, numOldLines-numToDelete)
			copy(newLines, oldLines[:len(newLines)])

			oldContent := strings.Join(oldLines, "\n")
			newContent := strings.Join(newLines, "\n")

			oldSnap := &files.Snapshot{
				Title:   fmt.Sprintf("Random Deletion %d", i),
				Test:    fmt.Sprintf("TestRandomDel_%d", i),
				Content: oldContent,
			}

			newSnap := &files.Snapshot{
				Title:   fmt.Sprintf("Random Deletion %d", i),
				Test:    fmt.Sprintf("TestRandomDel_%d", i),
				Content: newContent,
			}

			diffLines := diff.Histogram(oldContent, newContent)
			result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

			// Validate structure
			stripped := stripANSI(result)

			if !strings.Contains(stripped, "┬") {
				t.Error("Missing top bar")
			}
			if !strings.Contains(stripped, "┴") {
				t.Error("Missing bottom bar")
			}

			// Count deletions (at least numToDelete should appear)
			delCount := strings.Count(result, "-")
			if delCount < numToDelete {
				t.Errorf("Expected at least %d deletions, got %d", numToDelete, delCount)
			}
		})
	}
}

func TestDiffSnapshotBox_Random_Mixed(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Setenv("COLUMNS", "140")
	defer os.Unsetenv("COLUMNS")

	rng := rand.New(rand.NewSource(99999))

	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("random_mixed_%d", i), func(t *testing.T) {
			// Generate random old content (10-30 lines)
			numOldLines := rng.Intn(21) + 10
			oldLines := make([]string, numOldLines)
			for j := 0; j < numOldLines; j++ {
				oldLines[j] = fmt.Sprintf("old_line_%d_%s", j+1, randomWord(rng))
			}

			// Randomly modify, add, delete
			newLines := make([]string, 0, numOldLines*2)
			for j := 0; j < numOldLines; j++ {
				action := rng.Intn(100)
				if action < 70 { // 70% keep unchanged
					newLines = append(newLines, oldLines[j])
				} else if action < 85 { // 15% modify
					newLines = append(newLines, fmt.Sprintf("modified_%d_%s", j+1, randomWord(rng)))
				} else if action < 95 { // 10% add
					newLines = append(newLines, oldLines[j])
					newLines = append(newLines, fmt.Sprintf("added_%d_%s", j+1, randomWord(rng)))
				}
				// 5% delete (skip adding line)
			}

			oldContent := strings.Join(oldLines, "\n")
			newContent := strings.Join(newLines, "\n")

			oldSnap := &files.Snapshot{
				Title:   fmt.Sprintf("Random Mixed %d", i),
				Test:    fmt.Sprintf("TestRandomMixed_%d", i),
				Content: oldContent,
			}

			newSnap := &files.Snapshot{
				Title:   fmt.Sprintf("Random Mixed %d", i),
				Test:    fmt.Sprintf("TestRandomMixed_%d", i),
				Content: newContent,
			}

			diffLines := diff.Histogram(oldContent, newContent)
			result := pretty.DiffSnapshotBox(oldSnap, newSnap, diffLines)

			// Validate basic structure
			stripped := stripANSI(result)

			if !strings.Contains(stripped, "┬") {
				t.Error("Missing top bar")
			}
			if !strings.Contains(stripped, "┴") {
				t.Error("Missing bottom bar")
			}
			if !strings.Contains(stripped, fmt.Sprintf("Random Mixed %d", i)) {
				t.Error("Missing title")
			}

			// Should have some diff markers
			hasPlus := strings.Contains(result, "+")
			hasMinus := strings.Contains(result, "-")
			hasPipe := strings.Contains(result, "│")

			if !hasPlus && !hasMinus && !hasPipe {
				t.Error("Expected at least one type of diff marker")
			}
		})
	}
}

// Helper function for random word generation
func randomWord(rng *rand.Rand) string {
	words := []string{"apple", "banana", "cherry", "date", "elderberry", "fig", "grape", "honeydew"}
	return words[rng.Intn(len(words))]
}
