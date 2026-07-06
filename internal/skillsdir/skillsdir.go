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

// findSibling implements PRD §8 rule 2 — locate <repoDir>/skills next to the
// running binary, symlink-aware. This is the rule that makes a symlink install
// work: ~/.local/bin/skpp -> ~/projects/skpp/skpp resolves back to the repo.
//
// It is a thin entry that asks the OS for the running binary (os.Executable)
// and delegates the symlink/dir logic to resolveSiblingFromExe. os.Executable
// cannot be controlled in a test (it returns the test binary's own path), so
// the testable core lives in resolveSiblingFromExe.
//
// Returns (candidate, SourceSibling, true) on a hit; ("", 0, false) otherwise so
// Find() (S3) can fall through to rule 3. Never errors (locked per-rule shape).
func findSibling() (dir string, src Source, found bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", 0, false // no binary path at all -> skip rule
	}
	d, ok := resolveSiblingFromExe(exe)
	if !ok {
		return "", 0, false
	}
	return d, SourceSibling, true
}

// resolveSiblingFromExe is the symlink-aware sibling-resolution core for rule 2,
// factored out so it can be unit-tested with arbitrary exe paths.
//
// PRD §8.2 sequence:
//
//	real, err := filepath.EvalSymlinks(exe)  // REQUIRED on macOS (redundant but
//	                                         //   harmless on Linux via /proc/self/exe)
//	if err != nil { real = exe }             // fall back to raw exe on EvalSymlinks error
//	repoDir := filepath.Dir(real)
//	candidate := filepath.Join(repoDir, "skills")
//	win iff os.Stat(candidate) reports an existing directory
//
// See architecture/verified_symlink_resolution.md for why EvalSymlinks must stay.
func resolveSiblingFromExe(exe string) (dir string, found bool) {
	real, err := filepath.EvalSymlinks(exe)
	if err != nil {
		real = exe // EvalSymlinks could not resolve -> use exe verbatim
	}
	repoDir := filepath.Dir(real)
	candidate := filepath.Join(repoDir, "skills")
	info, err := os.Stat(candidate)
	if err != nil || !info.IsDir() {
		return "", false // no existing skills/ sibling -> rule misses
	}
	return candidate, true
}
