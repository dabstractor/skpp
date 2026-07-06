package skillsdir

import (
	"os"
	"path/filepath"
	"testing"
)

// unsetEnvVar removes envVar for the duration of the test and restores the
// prior state on cleanup. Needed because t.Setenv can only set, never unset.
// Do NOT call t.Parallel() in any test that uses this or t.Setenv.
func unsetEnvVar(tb testing.TB) {
	tb.Helper()
	prev, had := os.LookupEnv(envVar)
	if err := os.Unsetenv(envVar); err != nil {
		tb.Fatalf("unsetenv %s: %v", envVar, err)
	}
	tb.Cleanup(func() {
		if had {
			_ = os.Setenv(envVar, prev)
		} else {
			_ = os.Unsetenv(envVar)
		}
	})
}

func TestSourceString(t *testing.T) {
	cases := []struct {
		src  Source
		want string
	}{
		{SourceEnv, "SKPP_SKILLS_DIR"},
		{SourceSibling, "sibling of binary"},
		{SourceWalkUp, "ancestor of cwd"},
		{Source(-1), "unknown"}, // out-of-range -> default
		{Source(99), "unknown"},
	}
	for _, c := range cases {
		if got := c.src.String(); got != c.want {
			t.Errorf("Source(%d).String() = %q, want %q", c.src, got, c.want)
		}
	}
}

// Rule 1: env unset -> not found (fall through, no error).
func TestFindEnvUnset(t *testing.T) {
	unsetEnvVar(t)
	dir, src, found := findEnv()
	if found {
		t.Errorf("findEnv() env unset: got found=true dir=%q src=%v; want found=false", dir, src)
	}
}

// Rule 1: env set to "" -> not found (treated as no usable dir).
func TestFindEnvEmpty(t *testing.T) {
	t.Setenv(envVar, "")
	dir, src, found := findEnv()
	if found {
		t.Errorf("findEnv() env empty: got found=true dir=%q src=%v; want found=false", dir, src)
	}
}

// Rule 1: env set to an existing directory (absolute temp dir) -> found, abs path, SourceEnv.
func TestFindEnvExistingDir(t *testing.T) {
	dir := t.TempDir() // already absolute + clean
	t.Setenv(envVar, dir)
	got, src, found := findEnv()
	if !found {
		t.Fatalf("findEnv() existing dir: found=false; want true")
	}
	if src != SourceEnv {
		t.Errorf("findEnv() existing dir: src=%v; want SourceEnv", src)
	}
	if want := filepath.Clean(dir); got != want {
		t.Errorf("findEnv() existing dir: dir=%q; want %q", got, want)
	}
}

// Rule 1: env set to a path that does not exist -> not found (fall through).
func TestFindEnvNonexistent(t *testing.T) {
	t.Setenv(envVar, filepath.Join(t.TempDir(), "does-not-exist"))
	dir, src, found := findEnv()
	if found {
		t.Errorf("findEnv() nonexistent: got found=true dir=%q src=%v; want found=false", dir, src)
	}
}

// Rule 1: env set to a regular file (not a directory) -> not found (fall through).
func TestFindEnvRegularFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "afile")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(envVar, f)
	dir, src, found := findEnv()
	if found {
		t.Errorf("findEnv() regular file: got found=true dir=%q src=%v; want found=false", dir, src)
	}
}

// Rule 1: env set to a RELATIVE existing dir (".") -> found, absolutized via filepath.Abs.
// Proves the filepath.Abs (relative->absolute) path without chdir or cwd pollution.
func TestFindEnvRelativePathAbsolutized(t *testing.T) {
	t.Setenv(envVar, ".") // "." always exists and is a dir (the test's cwd)
	got, src, found := findEnv()
	if !found {
		t.Fatalf("findEnv() relative '.': found=false; want true")
	}
	if src != SourceEnv {
		t.Errorf("findEnv() relative '.': src=%v; want SourceEnv", src)
	}
	want, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("findEnv() relative '.': dir=%q; want absolutized %q", got, want)
	}
}

// Rule 1 CONTRACT: the env path must NOT be passed through EvalSymlinks.
// If SKPP_SKILLS_DIR points at a symlink-to-a-dir, findEnv must return the
// symlink path (made absolute/clean), NOT the resolved target. The user points
// exactly where they want.
func TestFindEnvDoesNotResolveSymlinks(t *testing.T) {
	realDir := t.TempDir()
	parent := t.TempDir()
	link := filepath.Join(parent, "link-to-skills")
	if err := os.Symlink(realDir, link); err != nil {
		t.Skipf("symlinks not supported on this platform: %v", err)
	}
	t.Setenv(envVar, link)
	got, src, found := findEnv()
	if !found {
		t.Fatalf("findEnv() symlink-to-dir: found=false; want true (os.Stat follows the symlink)")
	}
	if src != SourceEnv {
		t.Errorf("findEnv() symlink-to-dir: src=%v; want SourceEnv", src)
	}
	if got == realDir {
		t.Errorf("findEnv() symlink-to-dir: dir=%q == resolved target; must NOT EvalSymlinks the env path", got)
	}
	want, err := filepath.Abs(link)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("findEnv() symlink-to-dir: dir=%q; want symlink path (absolutized) %q", got, want)
	}
}
