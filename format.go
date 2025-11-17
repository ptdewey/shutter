package freeze

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
)

func TerminalWidth() int {
	width := os.Getenv("COLUMNS")
	if w, err := strconv.Atoi(width); err == nil && w > 0 {
		return w
	}
	return 80
}

func ClearScreen() {
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
}

func ClearLine() {
	fmt.Print("\033[K")
}

func Red(s string) string {
	if !hasColor() {
		return s
	}
	return colorRed + s + colorReset
}

func Green(s string) string {
	if !hasColor() {
		return s
	}
	return colorGreen + s + colorReset
}

func Yellow(s string) string {
	if !hasColor() {
		return s
	}
	return colorYellow + s + colorReset
}

func Blue(s string) string {
	if !hasColor() {
		return s
	}
	return colorBlue + s + colorReset
}

func Gray(s string) string {
	if !hasColor() {
		return s
	}
	return colorGray + s + colorReset
}

func Bold(s string) string {
	if !hasColor() {
		return s
	}
	return colorBold + s + colorReset
}

func hasColor() bool {
	return os.Getenv("NO_COLOR") == ""
}

func NewSnapshotBox(snap *Snapshot) string {
	width := TerminalWidth()

	var sb strings.Builder
	sb.WriteString(strings.Repeat("─", width+2) + "\n")
	// TODO: add file path to a new line below this
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

// TODO: this probably needs the styling overhaul from above
func DiffSnapshotBox(old, new *Snapshot) string {
	width := TerminalWidth()

	diffLines := Histogram(old.Content, new.Content)

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

func FormatHeader(text string) string {
	return Bold(Blue(text))
}

func FormatSuccess(text string) string {
	return Green(text)
}

func FormatError(text string) string {
	return Red(text)
}

func FormatWarning(text string) string {
	return Yellow(text)
}
