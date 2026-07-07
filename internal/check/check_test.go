package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dabstractor/skilldozer/internal/discover"
)

// mkSkill writes content to a temp skills/<relTag>/SKILL.md and returns the
// discover.Skill the way discover.Index would: parse (ignoring err, like Index),
// then discover.BuildSkill. Each skill gets its own temp root so SourceFile is
// unique — check re-parses each SourceFile independently, so isolation is correct.
// relTag uses '/' separators (cross-platform via filepath.FromSlash).
//
// This mirrors main_test.go's writeSkillTree but produces a single Skill, so
// check tests can build an arbitrary []discover.Skill (incl. dups) without a
// shared root.
func mkSkill(t *testing.T, relTag, content string) discover.Skill {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, filepath.FromSlash(relTag))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", dir, err)
	}
	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	fm, _, _ := discover.ParseFrontmatter(path) // Index ignores err; check re-parses
	return discover.BuildSkill(dir, relTag, fm)
}

// skill returns a Skill with a valid block, for the cases that only vary name/desc.
func skill(t *testing.T, relTag, name, desc string) discover.Skill {
	t.Helper()
	return mkSkill(t, relTag, "---\nname: "+name+"\ndescription: "+desc+"\n---\n# body\n")
}

// repeat returns a string of n copies of s (for boundary-length name/desc tests).
func repeat(s string, n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString(s)
	}
	return b.String()
}

func TestCheckValidSkillIsClean(t *testing.T) {
	rep := Check([]discover.Skill{skill(t, "example", "example", "A demo skill.")})
	if len(rep.BySkill) != 1 || len(rep.BySkill[0].Findings) != 0 {
		t.Errorf("clean skill should have zero findings; got %+v", rep.BySkill[0].Findings)
	}
	if rep.Errors != 0 || rep.Warnings != 0 {
		t.Errorf("clean skill: Errors=%d Warnings=%d; want 0,0", rep.Errors, rep.Warnings)
	}
	if rep.HasErrors() {
		t.Errorf("clean skill: HasErrors=true; want false")
	}
}

func TestCheckMissingFrontmatterBlock(t *testing.T) {
	// No '---' fences at all -> HasFM false, err nil -> ONE root-cause ERROR.
	s := mkSkill(t, "bare", "# just a heading\nno frontmatter here\n")
	rep := Check([]discover.Skill{s})
	fs := rep.BySkill[0].Findings
	if len(fs) != 1 || fs[0].Level != LevelError || !strings.Contains(fs[0].Message, "missing frontmatter block") {
		t.Errorf("no-block skill should be one 'missing frontmatter block' ERROR; got %+v", fs)
	}
	if rep.Errors != 1 {
		t.Errorf("Errors=%d; want 1", rep.Errors)
	}
}

func TestCheckMalformedYAML(t *testing.T) {
	// Broken YAML between valid fences -> ParseFrontmatter returns err.
	s := mkSkill(t, "broken", "---\nname: [unclosed\n---\n# body\n")
	rep := Check([]discover.Skill{s})
	fs := rep.BySkill[0].Findings
	if len(fs) != 1 || fs[0].Level != LevelError || !strings.Contains(fs[0].Message, "invalid SKILL.md frontmatter") {
		t.Errorf("malformed YAML should be one 'invalid SKILL.md frontmatter' ERROR; got %+v", fs)
	}
}

func TestCheckMissingName(t *testing.T) {
	s := mkSkill(t, "a", "---\ndescription: has desc but no name\n---\nx\n")
	rep := Check([]discover.Skill{s})
	fs := rep.BySkill[0].Findings
	if len(fs) != 1 || !strings.Contains(fs[0].Message, "'name' is missing") {
		t.Errorf("missing name -> one 'name is missing' ERROR; got %+v", fs)
	}
}

func TestCheckMissingDescription(t *testing.T) {
	s := mkSkill(t, "a", "---\nname: a\n---\nx\n")
	rep := Check([]discover.Skill{s})
	fs := rep.BySkill[0].Findings
	if len(fs) != 1 || !strings.Contains(fs[0].Message, "'description' is missing or empty") {
		t.Errorf("missing description -> one ERROR; got %+v", fs)
	}
}

func TestCheckEmptyDescription(t *testing.T) {
	// Whitespace-only description trims to "" -> ERROR (not a WARN).
	s := mkSkill(t, "a", "---\nname: a\ndescription: \"   \"\n---\nx\n")
	rep := Check([]discover.Skill{s})
	if rep.Errors != 1 || !strings.Contains(rep.BySkill[0].Findings[0].Message, "description") {
		t.Errorf("whitespace-only description -> one description ERROR; got %+v", rep.BySkill[0].Findings)
	}
}

func TestCheckNameLeadingHyphen(t *testing.T) {
	s := skill(t, "a", "-foo", "d")
	if rep := Check([]discover.Skill{s}); rep.Errors == 0 || !strings.Contains(rep.BySkill[0].Findings[0].Message, "lowercase a-z0-9") {
		t.Errorf("leading hyphen -> charset ERROR; got %+v", rep.BySkill[0].Findings)
	}
}

func TestCheckNameTrailingHyphen(t *testing.T) {
	s := skill(t, "a", "foo-", "d")
	if rep := Check([]discover.Skill{s}); rep.Errors == 0 {
		t.Errorf("trailing hyphen -> charset ERROR; got %+v", rep.BySkill[0].Findings)
	}
}

func TestCheckNameConsecutiveHyphens(t *testing.T) {
	s := skill(t, "a", "foo--bar", "d")
	if rep := Check([]discover.Skill{s}); rep.Errors == 0 {
		t.Errorf("consecutive hyphens -> charset ERROR; got %+v", rep.BySkill[0].Findings)
	}
}

func TestCheckNameUppercase(t *testing.T) {
	s := skill(t, "a", "Foo", "d")
	if rep := Check([]discover.Skill{s}); rep.Errors == 0 {
		t.Errorf("uppercase name -> charset ERROR; got %+v", rep.BySkill[0].Findings)
	}
}

func TestCheckNameTooLong(t *testing.T) {
	long := repeat("a", 65) // 65 chars, otherwise valid
	s := skill(t, "a", long, "d")
	rep := Check([]discover.Skill{s})
	if rep.Errors != 1 {
		t.Errorf("65-char name -> 1 ERROR; got %d: %+v", rep.Errors, rep.BySkill[0].Findings)
	}
	if !strings.Contains(rep.BySkill[0].Findings[0].Message, "65 chars (max 64)") {
		t.Errorf("too-long message should name the length; got %q", rep.BySkill[0].Findings[0].Message)
	}
}

func TestCheckNameAtLimitOK(t *testing.T) {
	at := repeat("a", 64) // exactly 64, valid -> OK
	s := skill(t, "a", at, "d")
	if rep := Check([]discover.Skill{s}); rep.Errors != 0 {
		t.Errorf("64-char valid name should be OK; got %+v", rep.BySkill[0].Findings)
	}
}

func TestCheckDescriptionTooLongWarns(t *testing.T) {
	long := repeat("x", 1025)
	s := skill(t, "a", "a", long)
	rep := Check([]discover.Skill{s})
	if rep.Errors != 0 {
		t.Errorf("over-long description is a WARN, not an ERROR; Errors=%d", rep.Errors)
	}
	if rep.Warnings != 1 || rep.HasErrors() {
		t.Errorf("expected 1 warning, no errors; got Warnings=%d HasErrors=%v", rep.Warnings, rep.HasErrors())
	}
	if !strings.Contains(rep.BySkill[0].Findings[0].Message, "1025 chars (max 1024)") {
		t.Errorf("WARN should name the length; got %q", rep.BySkill[0].Findings[0].Message)
	}
}

func TestCheckDescriptionAtLimitOK(t *testing.T) {
	at := repeat("x", 1024) // exactly 1024 -> no WARN
	s := skill(t, "a", "a", at)
	if rep := Check([]discover.Skill{s}); rep.Warnings != 0 || rep.Errors != 0 {
		t.Errorf("1024-char description should be clean; got W=%d E=%d", rep.Warnings, rep.Errors)
	}
}

func TestCheckDuplicateNames(t *testing.T) {
	a := skill(t, "alpha", "shared", "d")
	b := skill(t, "beta", "shared", "d")
	rep := Check([]discover.Skill{a, b})
	if rep.Errors != 2 {
		t.Errorf("two skills sharing a name -> 2 ERRORs; got %d", rep.Errors)
	}
}

func TestCheckDupMessageNamesOtherTag(t *testing.T) {
	a := skill(t, "alpha", "shared", "d")
	b := skill(t, "beta", "shared", "d")
	rep := Check([]discover.Skill{a, b})
	// alpha's dup ERROR must name beta (sorted "also in" list), and vice versa.
	alphaMsg := rep.BySkill[0].Findings[0].Message
	betaMsg := rep.BySkill[1].Findings[0].Message
	if !strings.Contains(alphaMsg, "beta") || !strings.Contains(alphaMsg, "duplicate") {
		t.Errorf("alpha dup message should name beta: %q", alphaMsg)
	}
	if !strings.Contains(betaMsg, "alpha") {
		t.Errorf("beta dup message should name alpha: %q", betaMsg)
	}
}

func TestCheckMissingNameNotCountedAsDup(t *testing.T) {
	// A skill with NO name must NOT participate in the dup scan (it has its own
	// missing-name ERROR). Here one skill has name "x", another has no name:
	// no dup ERROR should appear, only the single missing-name ERROR.
	withName := skill(t, "alpha", "x", "d")
	noName := mkSkill(t, "beta", "---\ndescription: d\n---\nx\n")
	rep := Check([]discover.Skill{withName, noName})
	for _, f := range rep.BySkill[0].Findings {
		if strings.Contains(f.Message, "duplicate") {
			t.Errorf("alpha should have NO dup ERROR (no other 'x'); got %q", f.Message)
		}
	}
	if rep.Errors != 1 { // only the missing-name ERROR on beta
		t.Errorf("expected exactly 1 ERROR (missing name on beta); got %d", rep.Errors)
	}
}

func TestCheckEmptyInputNoPanic(t *testing.T) {
	rep := Check(nil)
	if rep.Errors != 0 || rep.Warnings != 0 || len(rep.BySkill) != 0 {
		t.Errorf("Check(nil) should be empty clean report; got %+v", rep)
	}
	if rep.HasErrors() {
		t.Errorf("Check(nil).HasErrors()=true; want false")
	}
}
