package diff_test

import (
	"testing"

	"github.com/ptdewey/shutter/internal/diff"
)

func TestHistogramEmpty(t *testing.T) {
	both := diff.Histogram("", "")
	if both != nil {
		t.Errorf("empty strings should return nil, got %v", both)
	}

	oldEmpty := diff.Histogram("", "hello")
	if oldEmpty == nil {
		t.Errorf("new content should return non-nil")
	}

	newEmpty := diff.Histogram("hello", "")
	if newEmpty == nil {
		t.Errorf("old content should return non-nil")
	}
}

func TestHistogramIdentical(t *testing.T) {
	old := "line1\nline2\nline3"
	new := "line1\nline2\nline3"

	result := diff.Histogram(old, new)

	if len(result) != 3 {
		t.Errorf("expected 3 diff lines, got %d", len(result))
	}

	for i, dl := range result {
		if dl.Kind != diff.DiffShared {
			t.Errorf("line %d: expected DiffShared, got %v", i, dl.Kind)
		}
		expectedLineNum := i + 1
		if dl.OldNumber != expectedLineNum {
			t.Errorf("line %d: expected OldNumber=%d, got %d", i, expectedLineNum, dl.OldNumber)
		}
		if dl.NewNumber != expectedLineNum {
			t.Errorf("line %d: expected NewNumber=%d, got %d", i, expectedLineNum, dl.NewNumber)
		}
	}
}

func TestHistogramCompletelyDifferent(t *testing.T) {
	old := "old content"
	new := "new content"

	result := diff.Histogram(old, new)

	if len(result) != 2 {
		t.Errorf("expected 2 diff lines, got %d", len(result))
	}

	hasOld := false
	hasNew := false
	for _, dl := range result {
		if dl.Kind == diff.DiffOld {
			hasOld = true
		}
		if dl.Kind == diff.DiffNew {
			hasNew = true
		}
	}

	if !hasOld || !hasNew {
		t.Error("expected both old and new diff kinds")
	}
}

func TestHistogramSingleLineChange(t *testing.T) {
	old := "line1\nline2\nline3"
	new := "line1\nmodified\nline3"

	result := diff.Histogram(old, new)

	if len(result) < 3 {
		t.Errorf("expected at least 3 diff lines, got %d", len(result))
	}

	if result[0].Kind != diff.DiffShared || result[0].Line != "line1" {
		t.Errorf("line 0: expected shared 'line1', got %v %s", result[0].Kind, result[0].Line)
	}

	hasModified := false
	for _, dl := range result {
		if dl.Line == "modified" {
			hasModified = true
			if dl.Kind != diff.DiffNew {
				t.Errorf("'modified' should be marked as new, got %v", dl.Kind)
			}
		}
	}
	if !hasModified {
		t.Error("diff missing 'modified' line")
	}
}

func TestHistogramAddLine(t *testing.T) {
	old := "line1\nline2"
	new := "line1\nline1.5\nline2"

	result := diff.Histogram(old, new)

	newCount := 0
	for _, dl := range result {
		if dl.Kind == diff.DiffNew {
			newCount++
			if dl.Line != "line1.5" {
				t.Errorf("expected new line 'line1.5', got '%s'", dl.Line)
			}
		}
	}

	if newCount != 1 {
		t.Errorf("expected 1 new line, got %d", newCount)
	}
}

func TestHistogramRemoveLine(t *testing.T) {
	old := "line1\nline2\nline3"
	new := "line1\nline3"

	result := diff.Histogram(old, new)

	oldCount := 0
	for _, dl := range result {
		if dl.Kind == diff.DiffOld {
			oldCount++
			if dl.Line != "line2" {
				t.Errorf("expected old line 'line2', got '%s'", dl.Line)
			}
		}
	}

	if oldCount != 1 {
		t.Errorf("expected 1 old line, got %d", oldCount)
	}
}

func TestHistogramLineNumbers(t *testing.T) {
	old := "a\nb\nc"
	new := "a\nb\nc"

	result := diff.Histogram(old, new)

	for i, dl := range result {
		expectedLineNum := i + 1
		if dl.OldNumber != expectedLineNum {
			t.Errorf("line %d: expected OldNumber=%d, got %d", i, expectedLineNum, dl.OldNumber)
		}
		if dl.NewNumber != expectedLineNum {
			t.Errorf("line %d: expected NewNumber=%d, got %d", i, expectedLineNum, dl.NewNumber)
		}
	}
}

func TestHistogramMultilineChanges(t *testing.T) {
	old := "start\nmiddle\nend"
	new := "start\nnew1\nnew2\nend"

	result := diff.Histogram(old, new)

	newCount := 0
	for _, dl := range result {
		if dl.Kind == diff.DiffNew {
			newCount++
		}
	}

	if newCount != 2 {
		t.Errorf("expected 2 new lines, got %d", newCount)
	}
}

func TestHistogramWithEmptyLines(t *testing.T) {
	old := "line1\n\nline3"
	new := "line1\nline2\nline3"

	result := diff.Histogram(old, new)

	if len(result) == 0 {
		t.Error("expected non-empty diff result")
	}
}
