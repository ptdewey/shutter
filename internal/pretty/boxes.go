package pretty

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ptdewey/freeze/internal/files"
)

type DiffLine struct {
	Number int
	Line   string
	Kind   DiffKind
}

type DiffKind int

const (
	DiffShared DiffKind = iota
	DiffOld
	DiffNew
)

func NewSnapshotBox(snap *files.Snapshot) string {
	width := TerminalWidth()

	var sb strings.Builder
	sb.WriteString(strings.Repeat("─", width+2) + "\n")
	// TODO: "New Snapshot" should be above this line, in default color.
	// - color should be on test name and path
	sb.WriteString(fmt.Sprintf("  %s \n", Blue("New Snapshot -- \""+snap.TestName+"\"")))

	lines := strings.Split(snap.Content, "\n")
	numLines := len(lines)
	lineNumWidth := len(strconv.Itoa(numLines))

	topBar := strings.Repeat("─", lineNumWidth+3) + "┬" + strings.Repeat("─", width-lineNumWidth-2) + "\n"
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

	bottomBar := strings.Repeat("─", lineNumWidth+3) + "┴" + strings.Repeat("─", width-lineNumWidth-2) + "\n"
	sb.WriteString(bottomBar)

	return sb.String()
}

// TODO: needs to get overhauled with styling like above
// - should show line numbers, line numbers with diffs should be the same
// - should show test name and path in the header section
func DiffSnapshotBox(old, new *files.Snapshot, diffLines []DiffLine) string {
	width := TerminalWidth()

	var sb strings.Builder
	sb.WriteString(strings.Repeat("─", width) + "\n")
	sb.WriteString(fmt.Sprintf("  %s\n", Blue("Snapshot Diff")))
	sb.WriteString(strings.Repeat("─", width) + "\n")

	for _, dl := range diffLines {
		var prefix string
		var formatted string

		switch dl.Kind {
		case DiffOld:
			prefix = Red("−")
			formatted = Red(dl.Line)
		case DiffNew:
			prefix = Green("+")
			formatted = Green(dl.Line)
		case DiffShared:
			prefix = " "
			formatted = dl.Line
		}

		display := fmt.Sprintf("%s %s", prefix, formatted)
		if len(display) > width-4 {
			display = display[:width-7] + "..."
		}
		sb.WriteString(fmt.Sprintf("  %s\n", display))
	}

	sb.WriteString(strings.Repeat("─", width) + "\n")
	return sb.String()
}
