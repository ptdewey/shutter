package diff

/*
This file was sourced from github.com/gkampitakis/go-snaps, available with the following License:

MIT License

Copyright (c) 2021 Georgios Kampitakis

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

=======================

This package is a partial port of Python difflib.

This library is vendored and modified to address the `go-snaps` needs for a more
readable difference report.


Original source: https://github.com/pmezard/go-difflib

Copyright (c) 2013, Patrick Mezard
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

    Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
    Redistributions in binary form must reproduce the above copyright
notice, this list of conditions and the following disclaimer in the
documentation and/or other materials provided with the distribution.
    The names of its contributors may not be used to endorse or promote
products derived from this software without specific prior written
permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS
IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

import "strconv"

type DiffKind int

const (
	DiffShared DiffKind = iota
	DiffOld
	DiffNew
)

type DiffLine struct {
	OldNumber int
	NewNumber int
	Line      string
	Kind      DiffKind
}

const (
	// Tag Codes for internal opcodes
	opEqual int8 = iota
	opInsert
	opDelete
	opReplace
)

type match struct {
	A    int
	B    int
	Size int
}

type opCode struct {
	Tag int8
	I1  int
	I2  int
	J1  int
	J2  int
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// sequenceMatcher compares sequence of strings. The basic
// algorithm predates, and is a little fancier than, an algorithm
// published in the late 1980's by Ratcliff and Obershelp under the
// hyperbolic name "gestalt pattern matching".  The basic idea is to find
// the longest contiguous matching subsequence that contains no "junk"
// elements (R-O doesn't address junk).  The same idea is then applied
// recursively to the pieces of the sequences to the left and to the right
// of the matching subsequence.  This does not yield minimal edit
// sequences, but does tend to yield matches that "look right" to people.
//
// sequenceMatcher tries to compute a "human-friendly diff" between two
// sequences.  Unlike e.g. UNIX(tm) diff, the fundamental notion is the
// longest *contiguous* & junk-free matching subsequence.  That's what
// catches peoples' eyes.  The Windows(tm) windiff has another interesting
// notion, pairing up elements that appear uniquely in each sequence.
// That, and the method here, appear to yield more intuitive difference
// reports than does diff.  This method appears to be the least vulnerable
// to synching up on blocks of "junk lines", though (like blank lines in
// ordinary text files, or maybe "<P>" lines in HTML files).  That may be
// because this is the only method of the 3 that has a *concept* of
// "junk" <wink>.
//
// Timing:  Basic R-O is cubic time worst case and quadratic time expected
// case.  sequenceMatcher is quadratic time for the worst case and has
// expected-case behavior dependent in a complicated way on how many
// elements the sequences have in common; best case time is linear.
type sequenceMatcher struct {
	a              []string
	b              []string
	b2j            map[string][]int
	IsJunk         func(string) bool
	autoJunk       bool
	bJunk          map[string]struct{}
	matchingBlocks []match
	fullBCount     map[string]int
	bPopular       map[string]struct{}
	opCodes        []opCode
}

func newMatcher(a, b []string) *sequenceMatcher {
	m := sequenceMatcher{autoJunk: true}
	m.setSeqs(a, b)
	return &m
}

// Set two sequences to be compared.
func (m *sequenceMatcher) setSeqs(a, b []string) {
	m.setSeq1(a)
	m.setSeq2(b)
}

// Set the first sequence to be compared.
func (m *sequenceMatcher) setSeq1(a []string) {
	if &a == &m.a {
		return
	}
	m.a = a
	m.matchingBlocks, m.opCodes = nil, nil
}

// Set the second sequence to be compared.
func (m *sequenceMatcher) setSeq2(b []string) {
	if &b == &m.b {
		return
	}
	m.b = b
	m.matchingBlocks, m.opCodes, m.fullBCount = nil, nil, nil
	m.chainB()
}

func (m *sequenceMatcher) chainB() {
	// Populate line -> index mapping
	b2j := map[string][]int{}
	for i, elt := range m.b {
		indices := b2j[elt]
		indices = append(indices, i)
		b2j[elt] = indices
	}

	// Purge junk elements
	m.bJunk = map[string]struct{}{}
	if m.IsJunk != nil {
		junk := m.bJunk
		for elt := range b2j {
			if m.IsJunk(elt) {
				junk[elt] = struct{}{}
			}
		}
		for elt := range junk { // separate loop avoids separate list of keys
			delete(b2j, elt)
		}
	}

	// Purge popular elements that are not junk
	popular := map[string]struct{}{}
	n := len(m.b)
	if m.autoJunk && n >= 200 {
		ntest := n/100 + 1
		for s, indices := range b2j {
			if len(indices) > ntest {
				popular[s] = struct{}{}
			}
		}
		for s := range popular {
			delete(b2j, s)
		}
	}
	m.bPopular = popular
	m.b2j = b2j
}

func (m *sequenceMatcher) isBJunk(s string) bool {
	_, ok := m.bJunk[s]
	return ok
}

// Find longest matching block in a[alo:ahi] and b[blo:bhi].
func (m *sequenceMatcher) findLongestMatch(alo, ahi, blo, bhi int) match {
	besti, bestj, bestsize := alo, blo, 0

	// find longest junk-free match
	j2len := map[int]int{}
	for i := alo; i != ahi; i++ {
		newj2len := map[int]int{}
		for _, j := range m.b2j[m.a[i]] {
			if j < blo {
				continue
			}
			if j >= bhi {
				break
			}
			k := j2len[j-1] + 1
			newj2len[j] = k
			if k > bestsize {
				besti, bestj, bestsize = i-k+1, j-k+1, k
			}
		}
		j2len = newj2len
	}

	// Extend the best by non-junk elements on each end
	for besti > alo && bestj > blo && !m.isBJunk(m.b[bestj-1]) &&
		m.a[besti-1] == m.b[bestj-1] {
		besti, bestj, bestsize = besti-1, bestj-1, bestsize+1
	}
	for besti+bestsize < ahi && bestj+bestsize < bhi &&
		!m.isBJunk(m.b[bestj+bestsize]) &&
		m.a[besti+bestsize] == m.b[bestj+bestsize] {
		bestsize++
	}

	// Suck up junk on each side
	for besti > alo && bestj > blo && m.isBJunk(m.b[bestj-1]) &&
		m.a[besti-1] == m.b[bestj-1] {
		besti, bestj, bestsize = besti-1, bestj-1, bestsize+1
	}
	for besti+bestsize < ahi && bestj+bestsize < bhi &&
		m.isBJunk(m.b[bestj+bestsize]) &&
		m.a[besti+bestsize] == m.b[bestj+bestsize] {
		bestsize++
	}

	return match{A: besti, B: bestj, Size: bestsize}
}

// Return list of triples describing matching subsequences.
func (m *sequenceMatcher) getMatchingBlocks() []match {
	if m.matchingBlocks != nil {
		return m.matchingBlocks
	}

	var matchBlocks func(alo, ahi, blo, bhi int, matched []match) []match
	matchBlocks = func(alo, ahi, blo, bhi int, matched []match) []match {
		match := m.findLongestMatch(alo, ahi, blo, bhi)
		i, j, k := match.A, match.B, match.Size
		if match.Size > 0 {
			if alo < i && blo < j {
				matched = matchBlocks(alo, i, blo, j, matched)
			}
			matched = append(matched, match)
			if i+k < ahi && j+k < bhi {
				matched = matchBlocks(i+k, ahi, j+k, bhi, matched)
			}
		}
		return matched
	}
	matched := matchBlocks(0, len(m.a), 0, len(m.b), nil)

	// Collapse adjacent equal blocks
	nonAdjacent := []match{}
	i1, j1, k1 := 0, 0, 0
	for _, b := range matched {
		i2, j2, k2 := b.A, b.B, b.Size
		if i1+k1 == i2 && j1+k1 == j2 {
			k1 += k2
		} else {
			if k1 > 0 {
				nonAdjacent = append(nonAdjacent, match{i1, j1, k1})
			}
			i1, j1, k1 = i2, j2, k2
		}
	}
	if k1 > 0 {
		nonAdjacent = append(nonAdjacent, match{i1, j1, k1})
	}

	nonAdjacent = append(nonAdjacent, match{len(m.a), len(m.b), 0})
	m.matchingBlocks = nonAdjacent
	return m.matchingBlocks
}

// Return list of opcodes describing how to turn a into b.
func (m *sequenceMatcher) getOpCodes() []opCode {
	if m.opCodes != nil {
		return m.opCodes
	}
	i, j := 0, 0
	matching := m.getMatchingBlocks()
	opCodes := make([]opCode, 0, len(matching))
	for _, m := range matching {
		ai, bj, size := m.A, m.B, m.Size
		var tag int8 = 0
		if i < ai && j < bj {
			tag = opReplace
		} else if i < ai {
			tag = opDelete
		} else if j < bj {
			tag = opInsert
		}
		if tag > 0 {
			opCodes = append(opCodes, opCode{tag, i, ai, j, bj})
		}
		i, j = ai+size, bj+size
		if size > 0 {
			opCodes = append(opCodes, opCode{opEqual, ai, i, bj, j})
		}
	}
	m.opCodes = opCodes
	return m.opCodes
}

// Histogram computes a diff between two strings using the Ratcliff-Obershelp algorithm
func Histogram(old, new string) []DiffLine {
	oldLines := splitLines(old)
	newLines := splitLines(new)

	matcher := newMatcher(oldLines, newLines)
	opcodes := matcher.getOpCodes()

	var result []DiffLine

	for _, op := range opcodes {
		switch op.Tag {
		case opEqual:
			for i := op.I1; i < op.I2; i++ {
				newIdx := i + (op.J1 - op.I1)
				result = append(result, DiffLine{
					Line:      oldLines[i],
					Kind:      DiffShared,
					OldNumber: i + 1,
					NewNumber: newIdx + 1,
				})
			}
		case opDelete:
			for i := op.I1; i < op.I2; i++ {
				result = append(result, DiffLine{
					Line:      oldLines[i],
					Kind:      DiffOld,
					OldNumber: i + 1,
				})
			}
		case opInsert:
			for j := op.J1; j < op.J2; j++ {
				result = append(result, DiffLine{
					Line:      newLines[j],
					Kind:      DiffNew,
					NewNumber: j + 1,
				})
			}
		case opReplace:
			for i := op.I1; i < op.I2; i++ {
				result = append(result, DiffLine{
					Line:      oldLines[i],
					Kind:      DiffOld,
					OldNumber: i + 1,
				})
			}
			for j := op.J1; j < op.J2; j++ {
				result = append(result, DiffLine{
					Line:      newLines[j],
					Kind:      DiffNew,
					NewNumber: j + 1,
				})
			}
		}
	}

	return result
}

// splitLines splits a string by newlines, handling empty strings
func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	lines := splitLinesKeepNewline(s)
	return lines
}

// splitLinesKeepNewline splits on newlines but removes the newlines themselves
func splitLinesKeepNewline(s string) []string {
	var lines []string
	var current string

	for _, char := range s {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

// FormatRangeUnified converts range to the "ed" format for unified diffs
func FormatRangeUnified(start, stop int) string {
	beginning := start + 1 // lines start numbering with one
	length := stop - start

	if length == 1 {
		return strconv.Itoa(beginning)
	}
	if length == 0 {
		beginning--
	}

	return strconv.Itoa(beginning) + "," + strconv.Itoa(length)
}
