package diff

import (
	"strings"
)

type DiffKind int

const (
	DiffShared DiffKind = iota
	DiffOld
	DiffNew
)

type DiffLine struct {
	Number int
	Line   string
	Kind   DiffKind
}

func Histogram(old, new string) []DiffLine {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	if len(oldLines) == 1 && oldLines[0] == "" {
		oldLines = []string{}
	}
	if len(newLines) == 1 && newLines[0] == "" {
		newLines = []string{}
	}

	matrix := computeEditDistance(oldLines, newLines)
	return traceback(oldLines, newLines, matrix)
}

func computeEditDistance(old, new []string) [][]int {
	m, n := len(old), len(new)
	matrix := make([][]int, m+1)
	for i := range matrix {
		matrix[i] = make([]int, n+1)
	}

	for i := 0; i <= m; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= n; j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if old[i-1] == new[j-1] {
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				matrix[i][j] = 1 + minThree(matrix[i-1][j], matrix[i][j-1], matrix[i-1][j-1])
			}
		}
	}

	return matrix
}

func traceback(old, new []string, matrix [][]int) []DiffLine {
	var result []DiffLine
	i, j := len(old), len(new)

	for i > 0 || j > 0 {
		if i > 0 && j > 0 && old[i-1] == new[j-1] {
			result = append([]DiffLine{{Line: old[i-1], Kind: DiffShared}}, result...)
			i--
			j--
		} else if j > 0 && (i == 0 || matrix[i][j-1] < matrix[i-1][j]) {
			result = append([]DiffLine{{Line: new[j-1], Kind: DiffNew}}, result...)
			j--
		} else if i > 0 {
			result = append([]DiffLine{{Line: old[i-1], Kind: DiffOld}}, result...)
			i--
		} else {
			result = append([]DiffLine{{Line: new[j-1], Kind: DiffNew}}, result...)
			j--
		}
	}

	for idx := range result {
		result[idx].Number = idx + 1
	}

	return result
}

func minThree(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
