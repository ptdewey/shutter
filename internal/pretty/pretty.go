package pretty

import (
	"fmt"
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
