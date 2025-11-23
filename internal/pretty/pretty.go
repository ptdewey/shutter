package pretty

import (
	"os"
	"strconv"
)

const (
	colorRed    = "\033[91m" // Bright red (ANSI color 9)
	colorGreen  = "\033[92m" // Bright green (ANSI color 10)
	colorYellow = "\033[93m" // Bright yellow (ANSI color 11)
	colorBlue   = "\033[94m" // Bright blue (ANSI color 12)
	colorGray   = "\033[90m" // Bright black/gray (ANSI color 8)
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

func hasColor() bool {
	return os.Getenv("NO_COLOR") == ""
}

// colorize wraps text with the given color code
func colorize(s, code string) string {
	if !hasColor() {
		return s
	}
	return code + s + colorReset
}

func Red(s string) string {
	return colorize(s, colorRed)
}

func Green(s string) string {
	return colorize(s, colorGreen)
}

func Yellow(s string) string {
	return colorize(s, colorYellow)
}

func Blue(s string) string {
	return colorize(s, colorBlue)
}

func Gray(s string) string {
	return colorize(s, colorGray)
}

func Bold(s string) string {
	return colorize(s, colorBold)
}

func Header(text string) string {
	return Bold(Blue(text))
}

func Success(text string) string {
	return Green(text)
}

func Error(text string) string {
	return Red(text)
}

func Warning(text string) string {
	return Yellow(text)
}
