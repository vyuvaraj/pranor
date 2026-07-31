// Package dap implements a Debug Adapter Protocol proxy for the Serv language.
// It bridges VS Code (or any DAP client) to Delve — translating breakpoint
// positions and stack-frame line numbers between .pnr source coordinates and
// the generated Go file coordinates.
package import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// SourceMap is a bidirectional mapping between line numbers in the generated
// Go file (.build/<hash>/main.go) and the original .pnr source file.
//
// The codegen already emits "// .pnr line N" comments immediately before each
// top-level statement, so no compiler changes are required to build this map.
type SourceMap struct {
	// pnrFile is the absolute path of the original .pnr source file.
	pnrFile string
	// goFile is the absolute path of the generated Go file.
	goFile string

	// goToPnr maps a Go line → .pnr line number.
	goToPnr map[int]int
	// pnrToGo maps a .pnr line number → the first Go line that represents it.
	pnrToGo map[int]int

	// sortedGoLines is a sorted slice of all Go lines in the map (for nearest lookup).
	sortedGoLines []int
}

// ParseSourceMap reads the generated Go file at genGoPath and constructs a
// SourceMap by scanning for "// .pnr line N" comment markers emitted by the
// Serv codegen.
//
// Each such comment is assumed to precede the Go statement that corresponds to
// the referenced .pnr line, so the mapping recorded is:
//
//	(commentLine + 1) → pnrLine   (the statement is on the next line)
//	commentLine       → pnrLine   (also recorded for exact-comment hits)
func ParseSourceMap(genGoPath, pnrPath string) (*SourceMap, error) {
	f, err := os.Open(genGoPath)
	if err != nil {
		return nil, fmt.Errorf("sourcemap: open %q: %w", genGoPath, err)
	}
	defer f.Close()

	sm := &SourceMap{
		pnrFile: pnrPath,
		goFile:  genGoPath,
		goToPnr: make(map[int]int),
		pnrToGo: make(map[int]int),
	}

	scanner := bufio.NewScanner(f)
	goLine := 0
	for scanner.Scan() {
		goLine++
		line := strings.TrimSpace(scanner.Text())

		// Match "// .pnr line N" comments emitted by the Serv codegen.
		if strings.HasPrefix(line, "// .pnr line ") {
			rest := strings.TrimPrefix(line, "// .pnr line ")
			pnrLine, err := strconv.Atoi(strings.TrimSpace(rest))
			if err != nil {
				continue
			}
			// Record: the comment line itself maps to pnrLine.
			sm.goToPnr[goLine] = pnrLine
			// Record: the Go statement on the next line maps to pnrLine.
			sm.goToPnr[goLine+1] = pnrLine
			// Record the first Go line for this srv line (prefer earlier lines).
			if _, exists := sm.pnrToGo[pnrLine]; !exists {
				sm.pnrToGo[pnrLine] = goLine + 1
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("sourcemap: scan %q: %w", genGoPath, err)
	}

	// Build sorted list of Go lines for nearest-neighbour lookup.
	sm.sortedGoLines = make([]int, 0, len(sm.goToPnr))
	for gl := range sm.goToPnr {
		sm.sortedGoLines = append(sm.sortedGoLines, gl)
	}
	sort.Ints(sm.sortedGoLines)

	return sm, nil
}

// GoToPnr returns the .pnr line number that corresponds to the given Go line
// number. ok is false when no mapping exists.
func (sm *SourceMap) GoToPnr(goLine int) (pnrLine int, ok bool) {
	pnrLine, ok = sm.goToPnr[goLine]
	return
}

// SrvToGo returns the Go line number that best corresponds to the given .pnr
// line number. ok is false when the map is empty.
func (sm *SourceMap) PnrToGo(pnrLine int) (goLine int, ok bool) {
	if gl, exists := sm.pnrToGo[pnrLine]; exists {
		return gl, true
	}
	// No exact match — fall through to nearest-neighbour search.
	return sm.nearestGoLine(pnrLine)
}

// nearestGoLine finds the Go line whose associated .pnr line is closest to
// the requested pnrLine. This handles cases where a breakpoint is placed on
// a line that has no direct mapping (e.g. a blank line or comment in .pnr).
func (sm *SourceMap) nearestGoLine(pnrLine int) (goLine int, ok bool) {
	if len(sm.sortedGoLines) == 0 {
		return 0, false
	}

	bestGoLine := sm.sortedGoLines[0]
	bestDiff := abs(sm.goToPnr[bestGoLine] - pnrLine)

	for _, gl := range sm.sortedGoLines {
		diff := abs(sm.goToPnr[gl] - pnrLine)
		if diff < bestDiff {
			bestDiff = diff
			bestGoLine = gl
		}
	}
	return bestGoLine, true
}

// GoToPnrApprox returns the .pnr line number for the given Go line, or if no
// exact mapping exists, the closest preceding Go line that has a mapping.
func (sm *SourceMap) GoToPnrApprox(goLine int) (pnrLine int, ok bool) {
	if pnrLine, ok = sm.goToPnr[goLine]; ok {
		return pnrLine, true
	}
	if len(sm.sortedGoLines) == 0 {
		return 0, false
	}
	bestGoLine := -1
	for _, gl := range sm.sortedGoLines {
		if gl <= goLine {
			bestGoLine = gl
		} else {
			break
		}
	}
	if bestGoLine != -1 {
		return sm.goToPnr[bestGoLine], true
	}
	return sm.goToPnr[sm.sortedGoLines[0]], true
}

// GoFile returns the absolute path of the generated Go source file.
func (sm *SourceMap) GoFile() string { return sm.goFile }

// SrvFile returns the absolute path of the original .pnr source file.
func (sm *SourceMap) PnrFile() string { return sm.pnrFile }

// Len returns the number of Go lines that have a known .pnr mapping.
func (sm *SourceMap) Len() int { return len(sm.goToPnr) }

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
