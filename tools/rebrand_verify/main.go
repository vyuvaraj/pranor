// rebrand_verify is a Go replacement for run_verify_audit.py.
//
// Usage:
//
//	go run ./tools/rebrand_verify             # from pranor/ root
//	go run ./tools/rebrand_verify -root /path
//
// Exit code 0 → all checks pass.
// Exit code 1 → one or more checks failed.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ─── ANSI colours ──────────────────────────────────────────────────────────────

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	gray   = "\033[90m"
)

func col(color, s string) string { return color + s + reset }

// ─── Check definition ──────────────────────────────────────────────────────────

type check struct {
	category string
	name     string
	pattern  string          // main regex to find
	exclude  []string        // lines matching any of these are ignored
	warning  bool            // true → WARN instead of FAIL
}

// ─── Hit ───────────────────────────────────────────────────────────────────────

type hit struct {
	rel  string // relative path from root
	line int
	text string
}

// ─── Check catalogue (mirrors run_verify_audit.py) ────────────────────────────

var checks = []check{
	// Category 1: PascalCase component names
	{"Category 1", "ServGate", `ServGate`, []string{"PranorGate"}, false},
	{"Category 1", "ServStore", `ServStore`, []string{"PranorVault"}, false},
	{"Category 1", "ServQueue", `ServQueue`, []string{"PranorPulse"}, false},
	{"Category 1", "ServCron", `ServCron`, []string{"PranorChrono"}, false},
	{"Category 1", "ServAuth", `ServAuth`, []string{"PranorAuth"}, false},
	{"Category 1", "ServCache", `ServCache`, []string{"PranorCache"}, false},
	{"Category 1", "ServMesh", `ServMesh`, []string{"PranorMesh"}, false},
	{"Category 1", "ServTrace", `ServTrace`, []string{"PranorTrace"}, false},
	{"Category 1", "ServConsole", `ServConsole`, []string{"PranorConsole"}, false},
	{"Category 1", "ServPool", `ServPool`, []string{"PranorPool"}, false},
	{"Category 1", "ServMail", `ServMail`, []string{"PranorNotify"}, false},
	{"Category 1", "ServFlow", `ServFlow`, []string{"PranorFlow"}, false},
	{"Category 1", "ServCloud", `ServCloud`, []string{"PranorDeploy"}, false},
	{"Category 1", "ServTunnel", `ServTunnel`, []string{"PranorTunnel"}, false},
	{"Category 1", "ServShared", `ServShared`, []string{"PranorCore"}, false},
	{"Category 1", "ServLock", `ServLock`, []string{"PranorLock"}, false},
	{"Category 1", "ServSecret", `ServSecret`, []string{"PranorSecret"}, false},
	{"Category 1", "ServRegistry", `ServRegistry`, []string{"PranorHub"}, false},
	{"Category 1", "Serv-lang", `Serv-lang`, nil, false},
	{"Category 1", "Servverse", `Servverse|ServVerse|servverse`, nil, false},

	// Category 2: Lowercase service names
	{"Category 2", "servgate", `\bservgate\b`, []string{"service", "server", "observe", "reserved"}, false},
	{"Category 2", "servstore", `\bservstore\b`, nil, false},
	{"Category 2", "servqueue", `\bservqueue\b`, nil, false},
	{"Category 2", "servcron", `\bservcron\b`, nil, false},
	{"Category 2", "servauth", `\bservauth\b`, nil, false},
	{"Category 2", "servcache", `\bservcache\b`, nil, false},
	{"Category 2", "servmesh", `\bservmesh\b`, nil, false},
	{"Category 2", "servtrace", `\bservtrace\b`, nil, false},
	{"Category 2", "servconsole", `\bservconsole\b`, nil, false},
	{"Category 2", "servpool", `\bservpool\b`, nil, false},
	{"Category 2", "servmail", `\bservmail\b`, nil, false},
	{"Category 2", "servflow", `\bservflow\b`, nil, false},
	{"Category 2", "servcloud", `\bservcloud\b`, nil, false},
	{"Category 2", "servtunnel", `\bservtunnel\b`, nil, false},
	{"Category 2", "servregistry", `\bservregistry\b`, nil, false},
	{"Category 2", "servlockctl", `\bservlockctl\b`, nil, false},
	{"Category 2", "servsecretctl", `\bservsecretctl\b`, nil, false},

	// Category 3: Environment variable prefixes
	{"Category 3", "SERV_ prefix", `\bSERV_`, []string{"PRANOR_"}, false},
	{"Category 3", "SERVGATE_ prefix", `SERVGATE_`, nil, false},
	{"Category 3", "SERVQUEUE_ prefix", `SERVQUEUE_`, nil, false},
	{"Category 3", "SERVSTORE_ prefix", `SERVSTORE_`, nil, false},
	{"Category 3", "SERVAUTH_ prefix", `SERVAUTH_`, nil, false},
	{"Category 3", "SERVCACHE_ prefix", `SERVCACHE_`, nil, false},
	{"Category 3", "SERVTRACE_ prefix", `SERVTRACE_`, nil, false},
	{"Category 3", "SERVMESH_ prefix", `SERVMESH_`, nil, false},
	{"Category 3", "SERVVERSE_ prefix", `SERVVERSE_`, nil, false},
	{"Category 3", "SERVLOCK_ prefix", `SERVLOCK_`, nil, false},
	{"Category 3", "SERVSECRET_ prefix", `SERVSECRET_`, nil, false},

	// Category 4: Binaries & extensions
	{"Category 4", ".srv extension", `\.srv\b`, []string{"observe", "reserved", "conserv", "preserv", "dns.srv"}, false},
	{"Category 4", "serv.exe binary", `\bserv\.exe\b`, nil, false},
	{"Category 4", "serv-lsp binary", `\bserv-lsp\b`, nil, false},
	{"Category 4", "servd daemon", `\bservd\b`, []string{"observed", "reserved", "conserved", "preserved"}, false},

	// Category 5: URL schemes
	{"Category 5", "serv:// scheme", `serv://`, []string{`pranor://`, `observe://`, `reserved://`}, false},
	{"Category 5", "Pranor Gate:// broken scheme", `Pranor Gate://`, nil, false},
	{"Category 5", "Pranor Vault:// broken scheme", `Pranor Vault://`, nil, false},
	{"Category 5", "Pranor Pulse:// broken scheme", `Pranor Pulse://`, nil, false},

	// Category 6: Docker
	{"Category 6", "ghcr.io/vyuvaraj/serv", `ghcr\.io/vyuvaraj/serv`, nil, false},
	{"Category 6", "servverse-net network", `servverse-net`, nil, false},
	{"Category 6", `docker serv- prefix`, `docker.*"serv-`, nil, false},

	// Category 7: Repository references
	{"Category 7", "vyuvaraj/serv/", `vyuvaraj/serv/`, []string{"vyuvaraj/pranor"}, false},
	{"Category 7", "vyuvaraj/serv-ee", `vyuvaraj/serv-ee`, nil, false},
	{"Category 7", "packages/Serv", `packages/Serv`, nil, false},
	{"Category 7", "homebrew-serv", `homebrew-serv`, nil, false},
	{"Category 7", "scoop-serv", `scoop-serv`, nil, false},

	// Category 8: camelCase identifiers
	{"Category 8", "servGate", `\bservGate`, nil, false},
	{"Category 8", "servStore", `\bservStore`, nil, false},
	{"Category 8", "servQueue", `\bservQueue`, nil, false},
	{"Category 8", "servAuth", `\bservAuth`, nil, false},
	{"Category 8", "servCache", `\bservCache`, nil, false},
	{"Category 8", "servMesh", `\bservMesh`, nil, false},
	{"Category 8", "servCron", `\bservCron`, nil, false},
	{"Category 8", "servTrace", `\bservTrace`, nil, false},
	{"Category 8", "servFlow", `\bservFlow`, nil, false},
	{"Category 8", "servPool", `\bservPool`, nil, false},
	{"Category 8", "servMail", `\bservMail`, nil, false},
	{"Category 8", "servCloud", `\bservCloud`, nil, false},
	{"Category 8", "servTunnel", `\bservTunnel`, nil, false},

	// Category 9: Internal references
	{"Category 9", "_serv_migrations table", `_serv_migrations`, nil, false},
	{"Category 9", ".serv/ config dir", `\.serv/`, nil, false},
	{"Category 9", ".serv-build-cache", `\.serv-build-cache`, nil, false},
	{"Category 9", "serv-build module", `\bserv-build\b`, nil, false},
	{"Category 9", `"serv/runtime" import`, `"serv/runtime"`, nil, false},
}

// ─── File filter config ────────────────────────────────────────────────────────

var allowedExts = map[string]bool{
	".go": true, ".md": true, ".yml": true, ".yaml": true,
	".json": true, ".html": true, ".js": true, ".css": true,
	".py": true, ".ps1": true, ".sh": true, ".bat": true,
	".txt": true, ".xml": true, ".rb": true, ".toml": true,
	".iss": true, ".nuspec": true, ".pnr": true,
}

var ignoreDirs = map[string]bool{
	".git": true, "node_modules": true, ".build": true,
	"target": true, "dist": true, "vendor": true, ".gemini": true,
}

// skipPath returns true for the tool's own source directory.
func skipPath(path, root string) bool {
	rel, _ := filepath.Rel(root, path)
	return strings.HasPrefix(rel, "pranor/tools/rebrand_verify")
}

// skipped base names (the tool's own files + check_rebrand variants)
func skipFile(base string) bool {
	return strings.HasPrefix(base, "rebrand_verify") ||
		strings.HasPrefix(base, "check_rebrand") ||
		strings.HasPrefix(base, "run_verify")
}

// ─── Scanner ───────────────────────────────────────────────────────────────────

type result struct {
	check check
	hits  []hit
}

func scanCheck(root string, c check, mainRE *regexp.Regexp, exclREs []*regexp.Regexp) []hit {
	var hits []hit

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if ignoreDirs[d.Name()] || skipPath(path, root) {
				return filepath.SkipDir
			}
			return nil
		}
		base := d.Name()
		if !allowedExts[strings.ToLower(filepath.Ext(base))] || skipFile(base) {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		rel, _ := filepath.Rel(root, path)
		lineNo := 0
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 1<<20), 1<<20)
		for sc.Scan() {
			lineNo++
			line := sc.Text()

			// Skip large roadmap COMPLETED sections (mirrors Python logic)
			if strings.Contains(base, "UNIFIED_ROADMAP") &&
				(strings.Contains(base, "COMPLETED") || lineNo < 700) {
				continue
			}

			if !mainRE.MatchString(line) {
				continue
			}
			excluded := false
			for _, ex := range exclREs {
				if ex.MatchString(line) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
			text := line
			if len(text) > 120 {
				text = text[:120] + "…"
			}
			hits = append(hits, hit{rel: rel, line: lineNo, text: strings.TrimSpace(text)})
		}
		return nil
	})
	return hits
}

// ─── Main ──────────────────────────────────────────────────────────────────────

func main() {
	defaultRoot := filepath.Join(os.Getenv("HOME"), "workspace")
	rootFlag := flag.String("root", defaultRoot, "workspace root to scan")
	flag.Parse()
	root := *rootFlag

	sep := strings.Repeat("═", 70)
	fmt.Println()
	fmt.Println(col(cyan, sep))
	fmt.Println(col(cyan, "  PRANOR REBRAND VERIFICATION — GO EDITION"))
	fmt.Println(col(cyan, sep))
	fmt.Printf("  Root : %s\n", root)
	fmt.Println(col(cyan, sep))
	fmt.Println()

	// Compile all regexes up-front
	type compiled struct {
		main  *regexp.Regexp
		excl  []*regexp.Regexp
	}
	compiled_checks := make([]compiled, len(checks))
	for i, c := range checks {
		mainRE, err := regexp.Compile(c.pattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "bad pattern for %q: %v\n", c.name, err)
			os.Exit(2)
		}
		var exclREs []*regexp.Regexp
		for _, e := range c.exclude {
			re, err := regexp.Compile(e)
			if err != nil {
				fmt.Fprintf(os.Stderr, "bad exclude pattern %q: %v\n", e, err)
				os.Exit(2)
			}
			exclREs = append(exclREs, re)
		}
		compiled_checks[i] = compiled{mainRE, exclREs}
	}

	passed, failed, warned := 0, 0, 0
	lastCat := ""
	var failures []result

	for i, c := range checks {
		if c.category != lastCat {
			lastCat = c.category
			fmt.Printf("%s\n", col(cyan, "--- "+c.category+" ---"))
		}

		hits := scanCheck(root, c, compiled_checks[i].main, compiled_checks[i].excl)

		if len(hits) == 0 {
			fmt.Printf("  %s %s\n", col(green, "PASS"), c.name)
			passed++
		} else if c.warning {
			fmt.Printf("  %s %s (%d hits)\n", col(yellow, "WARN"), c.name, len(hits))
			warned++
			// show up to 5 samples
			for j, h := range hits {
				if j >= 5 { break }
				fmt.Printf("       %s\n", col(gray, fmt.Sprintf("%s:%d: %s", h.rel, h.line, h.text)))
			}
			if len(hits) > 5 {
				fmt.Printf("       %s\n", col(gray, fmt.Sprintf("... and %d more", len(hits)-5)))
			}
		} else {
			fmt.Printf("  %s %s (%d hits)\n", col(red, "FAIL"), c.name, len(hits))
			failed++
			for j, h := range hits {
				if j >= 5 { break }
				fmt.Printf("       %s\n", col(gray, fmt.Sprintf("%s:%d: %s", h.rel, h.line, h.text)))
			}
			if len(hits) > 5 {
				fmt.Printf("       %s\n", col(gray, fmt.Sprintf("... and %d more", len(hits)-5)))
			}
			failures = append(failures, result{c, hits})
		}
	}

	fmt.Println()
	fmt.Println(col(cyan, sep))
	fmt.Printf("  Passed  : %s\n", col(green, fmt.Sprintf("%d", passed)))
	if warned > 0 {
		fmt.Printf("  Warnings: %s\n", col(yellow, fmt.Sprintf("%d", warned)))
	}
	if failed > 0 {
		fmt.Printf("  Failed  : %s\n", col(red, fmt.Sprintf("%d", failed)))
	} else {
		fmt.Printf("  Failed  : %s\n", col(green, "0"))
	}
	fmt.Printf("  Total   : %d\n", passed+warned+failed)
	fmt.Println(col(cyan, sep))

	if failed > 0 {
		fmt.Println()
		fmt.Println(col(red, "FAILED CHECKS:"))
		for _, r := range failures {
			fmt.Printf("  [%s] %s — %s (%d hits)\n",
				col(red, "FAIL"), r.check.category, r.check.name, len(r.hits))
		}
		fmt.Println()
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(col(green, "  All rebrand checks passed ✓"))
	fmt.Println()
}
