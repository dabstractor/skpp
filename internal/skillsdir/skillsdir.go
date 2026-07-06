// Package skillsdir locates the on-disk skills/ directory for skpp.
//
// It implements the PRD §8 priority order:
//
//  1. SKPP_SKILLS_DIR env var — if set and an existing dir, use it as-is.
//  2. Sibling of the running binary (symlink-aware via os.Executable + EvalSymlinks).
//  3. Walk up from the current working directory.
//
// The public entry point is Find() (added in P1.M1.T2.S3), which calls the
// per-rule helpers in order and returns the first hit. Each per-rule helper
// returns (dir string, src Source, found bool) where found is true only when
// that rule produced a usable absolute directory; on found==false the src value
// is meaningless and the caller falls through to the next rule.
package skillsdir

import (
	"os"
	"path/filepath"
)

// Source identifies which §8 rule located the skills directory. It is reported
// by `skpp --path` so users can tell how the dir was found.
type Source int

const (
	// SourceEnv means SKPP_SKILLS_DIR was set and pointed at an existing dir.
	SourceEnv Source = iota
	// SourceSibling means the skills dir was found next to the running binary.
	SourceSibling
	// SourceWalkUp means the skills dir was found by walking up from cwd.
	SourceWalkUp
)

// String returns a human-readable label for the rule that won, used by
// `skpp --path` reporting. Satisfies fmt.Stringer.
func (s Source) String() string {
	switch s {
	case SourceEnv:
		return "SKPP_SKILLS_DIR"
	case SourceSibling:
		return "sibling of binary"
	case SourceWalkUp:
		return "ancestor of cwd"
	default:
		return "unknown"
	}
}

// envVar is the environment variable consulted by rule 1. It is a package
// constant (not a parameter): the contract is "mock/replace nothing" — tests
// drive it via t.Setenv / os.Unsetenv, never via injection.
const envVar = "SKPP_SKILLS_DIR"

// findEnv implements PRD §8 rule 1.
//
// It reads SKPP_SKILLS_DIR; if the value names an existing directory it returns
// that directory as an absolute path with src=SourceEnv and found=true. The env
// path is NOT passed through filepath.EvalSymlinks: the user points exactly
// where they want (a symlink is preserved verbatim, only made absolute/clean
// via filepath.Abs). If the variable is unset, empty, or does not name an
// existing directory, it returns found=false with src's zero value so Find()
// can fall through to rule 2 — a bad env value never hard-errors.
func findEnv() (dir string, src Source, found bool) {
	val, ok := os.LookupEnv(envVar)
	if !ok || val == "" {
		return "", 0, false
	}
	info, err := os.Stat(val)
	if err != nil || !info.IsDir() {
		return "", 0, false // not an existing dir -> let the next rule try
	}
	abs, err := filepath.Abs(val)
	if err != nil {
		return "", 0, false // cwd unresolvable -> let the next rule try
	}
	return abs, SourceEnv, true
}
