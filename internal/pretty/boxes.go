package pretty

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ptdewey/freeze/internal/files"
)

type DiffLine struct {
	OldNumber int
	NewNumber int
	Line      string
	Kind      DiffKind
}

type DiffKind int

const (
	DiffShared DiffKind = iota
	DiffOld
	DiffNew
)

func NewSnapshotBox(snap *files.Snapshot) string {
	return newSnapshotBoxInternal(snap)
}

func DiffSnapshotBox(old, new *files.Snapshot, diffLines []DiffLine) string {
	width := TerminalWidth()
	snapshotFileName := files.SnapshotFileName(new.Test) + ".snap"

	var sb strings.Builder
	sb.WriteString(strings.Repeat("─", width) + "\n")
	// TODO: maybe make helper functions for this, swap coloring between the key and the value
	// TODO: maybe show the snapshot file name in gray next to the "a/r/s" options
	// (i.e. "a accept -> snap_file_name.snap", "reject" w/strikethrough?, skip, keeps "*snap.new")
	sb.WriteString(fmt.Sprintf("  file: %s\n", Gray(snapshotFileName)))
	sb.WriteString(fmt.Sprintf("  %s\n", Blue("Snapshot Diff")))
	if new.Title != "" {
		sb.WriteString(fmt.Sprintf("  title: %s\n", Blue("\""+new.Title+"\"")))
	}
	sb.WriteString(fmt.Sprintf("  test: %s\n", Blue("\""+new.Test+"\"")))
	sb.WriteString(strings.Repeat("─", width) + "\n")

	// Calculate max line numbers for proper spacing
	maxOldNum, maxNewNum := 0, 0
	for _, dl := range diffLines {
		if dl.OldNumber > maxOldNum {
			maxOldNum = dl.OldNumber
		}
		if dl.NewNumber > maxNewNum {
			maxNewNum = dl.NewNumber
		}
	}
	oldWidth := len(fmt.Sprintf("%d", maxOldNum))
	newWidth := len(fmt.Sprintf("%d", maxNewNum))

	for _, dl := range diffLines {
		var oldNumStr, newNumStr string
		var prefix string
		var formatted string

		switch dl.Kind {
		case DiffOld:
			oldNumStr = fmt.Sprintf("%*d", oldWidth, dl.OldNumber)
			newNumStr = strings.Repeat(" ", newWidth)
			prefix = Red("−")
			formatted = Red(dl.Line)
		case DiffNew:
			oldNumStr = strings.Repeat(" ", oldWidth)
			newNumStr = fmt.Sprintf("%*d", newWidth, dl.NewNumber)
			prefix = Green("+")
			formatted = Green(dl.Line)
		case DiffShared:
			oldNumStr = fmt.Sprintf("%*d", oldWidth, dl.OldNumber)
			newNumStr = fmt.Sprintf("%*d", newWidth, dl.NewNumber)
			prefix = " "
			formatted = dl.Line
		}

		linePrefix := fmt.Sprintf("%s %s %s", Gray(oldNumStr), Gray(newNumStr), prefix)
		display := fmt.Sprintf("%s %s", linePrefix, formatted)

		// Adjust for actual display length considering ANSI codes
		if len(dl.Line) > width-oldWidth-newWidth-8 {
			formatted = formatted[:width-oldWidth-newWidth-11] + "..."
			display = fmt.Sprintf("%s %s", linePrefix, formatted)
		}

		sb.WriteString(fmt.Sprintf("  %s\n", display))
	}

	sb.WriteString(strings.Repeat("─", width) + "\n")
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
