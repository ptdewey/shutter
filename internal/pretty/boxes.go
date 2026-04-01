package pretty

import (
	"fmt"
	"strings"

	"github.com/ptdewey/shutter/internal/diff"
	"github.com/ptdewey/shutter/internal/files"
)

func NewSnapshotBox(snap *files.Snapshot, width ...int) string {
	w := TerminalWidth()
	if len(width) > 0 && width[0] > 0 {
		w = width[0]
	}
	return newSnapshotBoxInternal(snap, w)
}

// calculateLineNumWidth returns the width needed to display line numbers
func calculateLineNumWidth(maxLineNum int) int {
	return len(fmt.Sprintf("%d", maxLineNum))
}

// formatColoredLine applies color to a line based on diff kind
func formatColoredLine(line string, kind diff.DiffKind) string {
	switch kind {
	case diff.DiffOld:
		return Red(line)
	case diff.DiffNew:
		return Green(line)
	case diff.DiffShared:
		return line
	default:
		return line
	}
}

func DiffSnapshotBox(old, newSnapshot *files.Snapshot, diffLines []diff.DiffLine, widthOpt ...int) string {
	width := TerminalWidth()
	if len(widthOpt) > 0 && widthOpt[0] > 0 {
		width = widthOpt[0]
	}
	snapshotFileName := files.SnapshotFileName(newSnapshot.Title) + ".snap"

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
	lineNumWidth := calculateLineNumWidth(maxLineNum)

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

		// Wrap long lines instead of truncating
		// Account for: 2 spaces padding + 2 line number columns + 2 spaces between + prefix + space
		maxContentWidth := width - (lineNumWidth * 2) - 8
		if maxContentWidth < 20 {
			maxContentWidth = 20
		}

		if len(dl.Line) > maxContentWidth {
			// Emit wrapped chunks with proper gutter alignment
			line := dl.Line
			first := true
			for len(line) > 0 {
				chunk := line
				if len(chunk) > maxContentWidth {
					chunk = line[:maxContentWidth]
					line = line[maxContentWidth:]
				} else {
					line = ""
				}
				coloredChunk := formatColoredLine(chunk, dl.Kind)
				if first {
					display := fmt.Sprintf("%s %s %s %s", leftNum, rightNum, prefix, coloredChunk)
					sb.WriteString(fmt.Sprintf("  %s\n", display))
					first = false
				} else {
					pad := strings.Repeat(" ", lineNumWidth)
					display := fmt.Sprintf("%s %s %s %s", pad, pad, "│", coloredChunk)
					sb.WriteString(fmt.Sprintf("  %s\n", display))
				}
			}
		} else {
			display := fmt.Sprintf("%s %s %s %s", leftNum, rightNum, prefix, formatted)
			sb.WriteString(fmt.Sprintf("  %s\n", display))
		}
	}

	// Bottom bar with corner (account for both line number columns)
	bottomBar := strings.Repeat("─", (lineNumWidth*2)+4) + "┴" +
		strings.Repeat("─", width-(lineNumWidth*2)-1) + "\n"
	sb.WriteString(bottomBar)

	return sb.String()
}

func newSnapshotBoxInternal(snap *files.Snapshot, width int) string {

	var sb strings.Builder
	sb.WriteString("─── " + "New Snapshot " + strings.Repeat("─", width-15) + "\n\n")

	if snap.Title != "" {
		sb.WriteString(Blue("  title: ") + snap.Title + "\n")
	}
	if snap.Test != "" {
		sb.WriteString(Blue("  test: ") + snap.Test + "\n")
	}
	if snap.FileName != "" {
		sb.WriteString(Blue("  file: ") + snap.FileName + "\n")
	}
	sb.WriteString("\n")

	lines := strings.Split(snap.Content, "\n")
	numLines := len(lines)
	lineNumWidth := calculateLineNumWidth(numLines)

	topBar := strings.Repeat("─", lineNumWidth+3) + "┬" +
		strings.Repeat("─", width-lineNumWidth-2) + "\n"
	sb.WriteString(topBar)

	for i, line := range lines {
		lineNum := fmt.Sprintf("%*d", lineNumWidth, i+1)
		prefix := fmt.Sprintf("%s %s", Green(lineNum), Green("+"))

		maxContentWidth := width - lineNumWidth - 6
		if maxContentWidth < 20 {
			maxContentWidth = 20
		}

		if len(line) > maxContentWidth {
			remaining := line
			first := true
			for len(remaining) > 0 {
				chunk := remaining
				if len(chunk) > maxContentWidth {
					chunk = remaining[:maxContentWidth]
					remaining = remaining[maxContentWidth:]
				} else {
					remaining = ""
				}
				if first {
					display := fmt.Sprintf("%s %s", prefix, Green(chunk))
					sb.WriteString(fmt.Sprintf("  %s\n", display))
					first = false
				} else {
					pad := strings.Repeat(" ", lineNumWidth)
					display := fmt.Sprintf("%s %s %s", pad, "│", Green(chunk))
					sb.WriteString(fmt.Sprintf("  %s\n", display))
				}
			}
		} else {
			display := fmt.Sprintf("%s %s", prefix, Green(line))
			sb.WriteString(fmt.Sprintf("  %s\n", display))
		}
	}

	bottomBar := strings.Repeat("─", lineNumWidth+3) + "┴" +
		strings.Repeat("─", width-lineNumWidth-2) + "\n"
	sb.WriteString(bottomBar)

	return sb.String()
}
