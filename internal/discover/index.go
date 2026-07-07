// index.go implements discover.Index — the on-disk skills/ walk that ties S1's
// ParseFrontmatter (discover.go) and S2's BuildSkill (skill.go) into the []Skill
// the rest of skpp consumes (PRD §7.1). This is the P1.M2.T5.S1 deliverable.
// discover.go (S1) owns the frontmatter model/parser; skill.go (S2) owns the Skill
// type + metadata extraction; index.go (T5) owns the WalkDir scan, the relTag
// normalization, the sort, and the parse-error policy. It is the data source for
// T6 (--list), T7 (resolve), T9 (--search), and T10 (check).
package discover

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// Index walks the skills directory at skillsDir and returns every skill it
// contains, as a []Skill sorted by canonical tag (RelTag) for deterministic
// output. It implements PRD §7.1 discovery (manifest-free: the catalog is rebuilt
// from disk on every call — there is no index file).
//
// A "skill" is any directory that directly contains a SKILL.md file; nested skills
// count (skills/writing/reddit/SKILL.md is a skill whose RelTag is
// "writing/reddit"). relTag is the skill dir path relative to skillsDir, with OS
// separators normalized to '/' via filepath.ToSlash — so tags are "writing/reddit"
// on every platform (PRD §7.2 step 1; go_architecture.md "relTag normalization").
//
// skillsDir is made absolute first (filepath.Abs), so every Skill.Dir is an
// absolute path — the contract behind PRD §6.1 ("absolute path") and the §13
// acceptance gate (`case "$(./skpp example)" in /*)`). On the canonical absolute
// input (from skillsdir.Find) Abs is a no-op Clean.
//
// Error policy (the decision S2's PRP assigned to T5; see research/verified_facts.md §8):
//   - skillsDir missing, unreadable, or not a directory -> returned as the error.
//     (The caller, main, prints it to stderr and exits 1, PRD §6.4/§8.4.)
//   - A per-entry error (an unreadable subtree) is SKIPPED; the walk continues.
//   - Malformed YAML inside a SKILL.md does NOT abort the walk and is NOT
//     propagated: ParseFrontmatter returns (Frontmatter{}, body, err); Index
//     ignores err and builds a HasFM=false Skill via BuildSkill so the skill is
//     still resolvable by directory/basename (PRD §7.1). check (M4/T10) can
//     re-run ParseFrontmatter(s.SourceFile) to distinguish "malformed YAML" from
//     "no frontmatter block" (idempotent; no rework).
//
// filepath.WalkDir does NOT follow symlinked directories (stdlib default); a
// symlink to a skill dir is therefore not discovered. PRD §7.1 does not require
// following symlinks, and the default avoids cycles.
//
// An empty skills dir (no SKILL.md anywhere) yields a nil slice and a nil error;
// callers test with len() (e.g. --list exits 1 "if no skills found").
func Index(skillsDir string) ([]Skill, error) {
	root, err := filepath.Abs(skillsDir)
	if err != nil {
		return nil, err
	}
	// Stat-guard BEFORE WalkDir: a missing root is otherwise SWALLOWED. WalkDir
	// feeds the root's lstat error to the callback, and the per-entry
	// `if err != nil { return nil }` below would hide it -> (nil, nil). See
	// research/verified_facts.md Run 1 (the bug) vs Run 2 (the fix).
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New(root + ": not a directory")
	}

	var skills []Skill
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // per-entry unreadable (e.g. a chmod-000 subdir) -> skip, keep walking
		}
		// A skill is identified by the FILE entry "SKILL.md"; directories and any
		// other filename are walked past. d.IsDir() guards against a directory
		// literally named "SKILL.md".
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}
		skillDir := filepath.Dir(path)
		rel, rerr := filepath.Rel(root, skillDir)
		if rerr != nil {
			return nil // skillDir is always under root (found by walking it), so this is unreachable
		}
		relTag := filepath.ToSlash(rel)
		// Lenient parse: a malformed or frontmatter-less SKILL.md still yields a
		// resolvable Skill (HasFM=false). err is intentionally ignored here; see
		// the doc comment above for the policy + the M4 re-parse note.
		fm, _, _ := ParseFrontmatter(path)
		skills = append(skills, BuildSkill(skillDir, relTag, fm))
		return nil
	})
	if walkErr != nil {
		return skills, walkErr
	}
	// Deterministic output: sort by canonical tag (PRD §6.1 --all "sorted by tag").
	sort.Slice(skills, func(i, j int) bool { return skills[i].RelTag < skills[j].RelTag })
	return skills, nil
}
