// Package check validates every skill in a manifest-free store against the PRD §9
// rules and PRD §3 Agent Skills name rules. It is a FUNCTION over
// []discover.Skill (the pre-sorted catalog from discover.Index) that returns a
// structured Report; main.run (P1.M4.T10.S1) renders the report to stdout and
// maps Report.HasErrors() to the exit code.
//
// It mirrors internal/search (a function over []discover.Skill in its own
// internal/ package with its own _test.go): the validation concern stays
// isolated, independently unit-testable, and out of the thin main dispatcher.
//
// The non-obvious part: discover.Index DROPS the per-skill frontmatter parse
// error (a malformed-YAML SKILL.md still builds a HasFM=false Skill so it stays
// resolvable by directory). check therefore RE-RUNS discover.ParseFrontmatter on
// each s.SourceFile to recover the malformed-YAML-vs-no-block distinction — the
// exact re-parse internal/discover/index.go's doc comment already documents for
// "check (M4/T10)". The double parse is cheap (small files, small store) and
// idempotent (no rework in discover).
package check

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/dabstractor/skpp/internal/discover"
)

// Severity ranks a finding. OK < WARN < ERROR. OK is the implicit value for a
// skill with no findings (main prints an "OK" line); it is never carried by a
// Finding. Exported so main can switch on it if needed (it renders via String()).
type Severity int

const (
	LevelOK Severity = iota
	LevelWarn
	LevelError
)

// String renders a Severity as the 3-5 char status word main left-pads to width 5
// (`OK   `, `WARN `, `ERROR`). Mirrors resolve.MatchKind.String() / Source.String().
// An out-of-range value renders as "OK" (the zero value), which is safe.
func (s Severity) String() string {
	switch s {
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "OK"
	}
}

// Finding is one validation result line for a single skill. A skill with zero
// findings is OK (main emits one "OK" line); a skill with N findings emits N
// ERROR/WARN lines. Message is empty for OK (OK findings are never created).
type Finding struct {
	Level   Severity
	Message string
}

// SkillReport binds a skill to its findings. BySkill is in the input order
// (discover.Index sorts by RelTag), so the report is deterministic.
type SkillReport struct {
	Skill    discover.Skill
	Findings []Finding // empty => the skill is OK
}

// Report is the full check output. BySkill is in input order; Errors/Warnings are
// the totals across all findings (drive the summary line + exit code).
type Report struct {
	BySkill  []SkillReport
	Errors   int
	Warnings int
}

// HasErrors reports whether any ERROR finding exists. main maps this to the exit
// code (PRD §9: exit 1 if any ERROR). WARNs never affect it.
func (r Report) HasErrors() bool { return r.Errors > 0 }

// nameOwner pairs a skill's canonical tag with its frontmatter name for the
// duplicate-name scan. Only skills whose frontmatter parsed (HasFM, no err) AND
// whose name is non-empty are collected — the only names that can duplicate.
// Defined at PACKAGE scope (not inside Check) so it can be passed to
// appendDupFindings without an anonymous-struct assignability mismatch (a named
// type and an anonymous struct with identical fields are NOT the same type in Go).
type nameOwner struct {
	relTag string
	name   string
}

// validName enforces the PRD §3 Agent Skills name charset + structure: lowercase
// a-z0-9 with single hyphens, no leading/trailing/consecutive hyphens. It CANNOT
// express the 64-char max (checked separately) nor emptiness (a missing name is
// its own ERROR, handled before this regex is consulted). Verified live: accepts
// example/foo-helper/a/123/a-b-c; rejects -foo/foo-/foo--bar/Foo/foo_bar.
var validName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// PRD §3 / §9 limits. nameLenMax is the Agent Skills name ceiling; descLenMax is
// the description ceiling (PRD §3: "description max 1024 chars"). Both measured on
// ASCII / trimmed content respectively (see Check).
const (
	nameLenMax = 64
	descLenMax = 1024
)

// Check validates every skill in skills against the PRD §9 rules and returns a
// structured Report. It is the P1.M4.T10.S1 deliverable; main.run renders it.
//
// Algorithm (three passes):
//
//  1. PER-SKILL local checks: re-parse s.SourceFile (recover malformed-YAML vs
//     no-block), then — only when the frontmatter parsed — check name presence,
//     name validity (charset/structure/length), description presence, and
//     description length. Collect each non-empty name for the dup scan.
//  2. GLOBAL duplicate-name scan: any non-empty name owned by >1 skill yields one
//     ERROR per owner, naming the other relTag(s) (sorted).
//  3. TALLY errors/warnings across all findings.
//
// A skill whose re-parse failed (malformed YAML) or had no '---' block gets ONE
// root-cause ERROR (its name/description are definitionally absent; field checks
// are skipped to avoid noise). Field checks run only when fm.HasFM && err == nil.
//
// check does NOT scan for "directories that lack SKILL.md but look like skills":
// discover.Index only emits dirs that CONTAIN a SKILL.md, and a heuristic for the
// gap would false-positive on legitimate grouping dirs (research §2). The §9
// "empty besides SKILL.md" WARN is intentionally NOT implemented (research §3):
// the shipped example skill IS only SKILL.md, and enabling it would break the
// §13 acceptance ("reports the example as OK").
func Check(skills []discover.Skill) Report {
	var rep Report
	rep.BySkill = make([]SkillReport, len(skills))

	// owners collects (relTag, name) ONLY for skills whose frontmatter parsed
	// (HasFM, no err) and whose name is non-empty — the only names that can dup.
	// nameOwner is package-scoped (see above) so it is passable to appendDupFindings.
	var owners []nameOwner

	// Pass 1: per-skill local checks.
	for i := range skills {
		s := skills[i]
		var findings []Finding
		fm, _, perr := discover.ParseFrontmatter(s.SourceFile)
		switch {
		case perr != nil:
			// Malformed YAML between fences, OR the file vanished between Index and
			// check (race) -> ParseFrontmatter returns the os/yaml error. This is the
			// reframed §9 "skill dir has no SKILL.md": an UNUSABLE SKILL.md.
			findings = append(findings, Finding{LevelError, "invalid SKILL.md frontmatter: " + perr.Error()})
		case !fm.HasFM:
			// No '---' block at all. Root cause: name+description are absent. One ERROR.
			findings = append(findings, Finding{LevelError, "missing frontmatter block (no '---' delimiters)"})
		default:
			findings = append(findings, checkFields(fm)...)
			if fm.Name != "" {
				owners = append(owners, nameOwner{s.RelTag, fm.Name})
			}
		}
		rep.BySkill[i] = SkillReport{Skill: s, Findings: findings}
	}

	// Pass 2: global duplicate-name scan (case-sensitive on non-empty names).
	appendDupFindings(&rep, owners)

	// Pass 3: tally.
	for i := range rep.BySkill {
		for _, f := range rep.BySkill[i].Findings {
			switch f.Level {
			case LevelError:
				rep.Errors++
			case LevelWarn:
				rep.Warnings++
			}
		}
	}
	return rep
}

// checkFields runs the per-field ERROR/WARN checks for a skill whose frontmatter
// parsed (HasFM, no err). Order: name presence -> name length -> name charset ->
// description presence -> description length. Each failure appends its own Finding
// (a skill can accumulate several, e.g. invalid name + over-long description).
//
// Description length is measured on the TRIMMED value (strings.TrimSpace), matching
// ui.go's display length: a folded-scalar trailing newline does not count, and a
// whitespace-only description trims to "" -> "missing or empty" ERROR (not a WARN).
func checkFields(fm discover.Frontmatter) []Finding {
	var f []Finding

	// name presence + validity.
	if fm.Name == "" {
		f = append(f, Finding{LevelError, "frontmatter 'name' is missing"})
	} else if len(fm.Name) > nameLenMax {
		f = append(f, Finding{LevelError, fmt.Sprintf("frontmatter 'name' is %d chars (max %d)", len(fm.Name), nameLenMax)})
	} else if !validName.MatchString(fm.Name) {
		f = append(f, Finding{LevelError, "frontmatter 'name' must be lowercase a-z0-9 with single hyphens (no leading/trailing/consecutive hyphens)"})
	}

	// description presence + length.
	desc := strings.TrimSpace(fm.Description)
	if desc == "" {
		f = append(f, Finding{LevelError, "frontmatter 'description' is missing or empty"})
	} else if len(desc) > descLenMax {
		f = append(f, Finding{LevelWarn, fmt.Sprintf("description is %d chars (max %d)", len(desc), descLenMax)})
	}

	return f
}

// appendDupFindings adds a duplicate-name ERROR to every skill that shares a
// non-empty frontmatter name with at least one other skill. owners is the
// (relTag, name) list collected in pass 1. The "also in" list excludes the skill
// itself and is sorted for deterministic output (PRD §6.4-style stable reports).
//
// It mutates rep.BySkill in place by matching sr.Skill.RelTag (RelTag is unique
// per directory, so the match is unambiguous). A skill whose name is invalid (bad
// charset) but duplicated still counts — the literal name string matches.
func appendDupFindings(rep *Report, owners []nameOwner) {
	byName := map[string][]string{}
	for _, o := range owners {
		byName[o.name] = append(byName[o.name], o.relTag)
	}
	for name, tags := range byName {
		if len(tags) < 2 {
			continue
		}
		sort.Strings(tags)
		for i := range rep.BySkill {
			sr := &rep.BySkill[i]
			if sr.Skill.Name != name {
				continue
			}
			others := make([]string, 0, len(tags)-1)
			for _, t := range tags {
				if t != sr.Skill.RelTag {
					others = append(others, t)
				}
			}
			sr.Findings = append(sr.Findings, Finding{
				Level:   LevelError,
				Message: fmt.Sprintf("duplicate frontmatter 'name' %q (also in: %s)", name, strings.Join(others, ", ")),
			})
		}
	}
}
