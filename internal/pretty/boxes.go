package pretty

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ptdewey/shutter/internal/diff"
	"github.com/ptdewey/shutter/internal/files"
)

func NewSnapshotBox(snap *files.Snapshot) string {
	return newSnapshotBoxInternal(snap)
}

func DiffSnapshotBox(old, newSnapshot *files.Snapshot, diffLines []diff.DiffLine) string {
	width := TerminalWidth()
	snapshotFileName := files.SnapshotFileName(newSnapshot.Test) + ".snap"

	var sb strings.Builder
	sb.WriteString("─── " + "Snapshot Diff " + strings.Repeat("─", width-15) + "\n\n")

	// TODO: maybe make helper functions for this, swap coloring between the key and the value
	// TODO: maybe show the snapshot file name in gray next to the "a/r/s" options
	// (i.e. "a accept -> snap_file_name.snap", "reject" w/strikethrough?, skip, keeps "*snap.new")
	if newSnapshot.Title != "" {
		sb.WriteString(Blue("  title: ") + newSnapshot.Title + "\n")
	}
	sb.WriteString(Blue("  test: ") + newSnapshot.Test + "\n")
	sb.WriteString(Blue("  file: ") + snapshotFileName + "\n")
	sb.WriteString("\n")
	// sb.WriteString(Red("  - old snapshot\n"))
	// sb.WriteString(Green("  + new snapshot\n"))
	// sb.WriteString("\n")

	// Calculate max line numbers for proper spacing
	maxOldNum := 0
	maxNewNum := 0
	for _, dl := range diffLines {
		if dl.OldNumber > maxOldNum {
			maxOldNum = dl.OldNumber
		}
		if dl.NewNumber > maxNewNum {
			maxNewNum = dl.NewNumber
		}
	}
	// Use the larger of the two for consistent column width
	maxLineNum := maxOldNum
	if maxNewNum > maxLineNum {
		maxLineNum = maxNewNum
	}
	lineNumWidth := len(fmt.Sprintf("%d", maxLineNum))

	// Top bar with corner (account for both line number columns)
	topBar := strings.Repeat("─", (lineNumWidth*2)+4) + "┬" +
		strings.Repeat("─", width-(lineNumWidth*2)-1) + "\n"
	sb.WriteString(topBar)

	for _, dl := range diffLines {
		var leftNum, rightNum, prefix, formatted string

		// FIX: line number coloring is the same between old and new lines
		switch dl.Kind {
		case diff.DiffOld:
			// For removed lines: show old line number on left, space on right, red -
			leftNum = Red(fmt.Sprintf("%*d", lineNumWidth, dl.OldNumber))
			rightNum = strings.Repeat(" ", lineNumWidth)
			prefix = Red("-")
			formatted = Red(dl.Line)
		case diff.DiffNew:
			// For added lines: space on left, new line number on right, green +
			leftNum = strings.Repeat(" ", lineNumWidth)
			rightNum = Green(fmt.Sprintf("%*d", lineNumWidth, dl.NewNumber))
			prefix = Green("+")
			formatted = Green(dl.Line)
		case diff.DiffShared:
			// For shared lines: show line number centered, │ separator (not gray)
			leftNum = strings.Repeat(" ", lineNumWidth)
			rightNum = Gray(fmt.Sprintf("%*d", lineNumWidth, dl.NewNumber))
			prefix = "│"
			formatted = dl.Line
		}

		// Adjust for actual display length considering ANSI codes
		// Account for: 2 spaces padding + 2 line number columns + 2 spaces between + prefix + space
		maxContentWidth := width - (lineNumWidth * 2) - 8
		if len(dl.Line) > maxContentWidth {
			truncated := dl.Line[:maxContentWidth-3] + "..."
			switch dl.Kind {
			case diff.DiffOld:
				formatted = Red(truncated)
			case diff.DiffNew:
				formatted = Green(truncated)
			case diff.DiffShared:
				formatted = truncated
			}
		}

		display := fmt.Sprintf("%s %s %s %s", leftNum, rightNum, prefix, formatted)
		sb.WriteString(fmt.Sprintf("  %s\n", display))
	}

	// Bottom bar with corner (account for both line number columns)
	bottomBar := strings.Repeat("─", (lineNumWidth*2)+4) + "┴" +
		strings.Repeat("─", width-(lineNumWidth*2)-1) + "\n"
	sb.WriteString(bottomBar)

	return sb.String()
}

func newSnapshotBoxInternal(snap *files.Snapshot) string {
	width := TerminalWidth()

	var sb strings.Builder
	sb.WriteString("─── " + "New Snapshot " + strings.Repeat("─", width-15) + "\n\n")

	if snap.Title != "" {
		sb.WriteString(Blue("  title: ") + snap.Title + "\n")
		// sb.WriteString(fmt.Sprintf("  title: %s\n", Blue(snap.Title)))
	}
	if snap.Test != "" {
		// sb.WriteString(fmt.Sprintf("  test: %s\n", Blue(snap.Test)))
		sb.WriteString(Blue("  test: ") + snap.Test + "\n")
	}
	if snap.FileName != "" {
		// sb.WriteString(fmt.Sprintf("  file: %s\n", Gray(snap.FileName)))
		sb.WriteString(Blue("  file: ") + snap.FileName + "\n")
	}
	sb.WriteString("\n")

	lines := strings.Split(snap.Content, "\n")
	numLines := len(lines)
	lineNumWidth := len(strconv.Itoa(numLines))

	topBar := strings.Repeat("─", lineNumWidth+3) + "┬" +
		strings.Repeat("─", width-lineNumWidth-2) + "\n"
	sb.WriteString(topBar)

	for i, line := range lines {
		lineNum := fmt.Sprintf("%*d", lineNumWidth, i+1)
		prefix := fmt.Sprintf("%s %s", Green(lineNum), Green("+"))

		if len(line) > width-len(prefix)-4 {
			line = line[:width-len(prefix)-7] + "..."
		}

		display := fmt.Sprintf("%s %s", prefix, Green(line))
		sb.WriteString(fmt.Sprintf("  %s\n", display))
	}

	bottomBar := strings.Repeat("─", lineNumWidth+3) + "┴" +
		strings.Repeat("─", width-lineNumWidth-2) + "\n"
	sb.WriteString(bottomBar)

	return sb.String()
}
