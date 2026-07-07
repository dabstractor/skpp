# Bug Fix Requirements

## Overview

End-to-end QA of the `skpp` implementation against the PRD. Testing covered the
full §13 acceptance suite (all pass), plus adversarial testing of tag resolution
precedence, discovery rules, error/atomicity contracts, frontmatter parsing,
shell completions, install.sh, unicode, permissions, and CLI flag handling.

**Overall assessment:** the implementation is solid. Every core contract holds:
path resolution, exit codes, stdout discipline (nothing on failure), atomicity
for multi-tag invocations, discovery priority, validation, and the pi
`--no-skills --skill` end-to-end load all work. Latency is ~3ms median, well
under the §4 <5ms target. Tests pass, `go vet` and `gofmt` are clean.

One Major issue and several Minor issues were found. None are Critical. The
Major issue is a documented diagnostic feature (`--path` should report which
discovery rule won) that is computed internally but never surfaced to the user.

## Critical Issues (Must Fix)

None. All core functionality works and every §13 acceptance criterion passes.

## Major Issues (Should Fix)

### Issue 1: `--path` does not report which discovery rule won
**Severity**: Major
**PRD Reference**: §8 ("`skpp --path` reports which rule won. This is the single
most failure-prone area") and §6.1 (`--path` row).
**Expected Behavior**: Per §8, `skpp --path` should tell the user which of the
three discovery rules (env var, sibling-of-binary, walk-up) produced the result,
because discovery is "the single most failure-prone area."
**Actual Behavior**: `--path` prints only the resolved directory path and nothing
about the source rule. The `Source` value is computed by `skillsdir.Find()` but
discarded in `main.go` (`dir, _, err := skillsdir.Find() // src is for reporting
only; not printed`). The entire `Source` enum, its `String()` method, and its
unit tests exist, but the value is assigned to `_` and the `String()` method is
dead-code-eliminated from the release binary (confirmed via `strings ./skpp`: the
labels "sibling of binary" / "ancestor of cwd" are absent).
**Steps to Reproduce**:
```bash
cd ~/projects/skpp
./skpp --path            # prints only /home/.../skills — no rule reported
# Compare with a typo'd env var that silently falls through to the sibling rule:
SKPP_SKILLS_DIR=/typo/not/real ./skpp --path   # prints repo skills/; indistinguishable from a real env hit
```
**Why it matters**: A user who typos `SKPP_SKILLS_DIR` gets silent fall-through to
the sibling/walk-up rule and sees a valid-looking path; there is no way to tell
the env var was ignored. This is exactly the failure-prone scenario §8 calls out.
**Suggested Fix**: Print the path to stdout (preserving the §13
`test "$(./skpp --path)" = "$PWD/skills"` contract) AND print the source label to
stderr, e.g. `fmt.Fprintf(stderr, "(found via %s)\n", src)`. The `Source.String()`
labels ("SKPP_SKILLS_DIR" / "sibling of binary" / "ancestor of cwd") already exist
and are unit-tested; they only need wiring into the `--path` branch.

## Minor Issues (Nice to Fix)

### Issue 2: Multi-byte (unicode) tags misalign the `--list` / `--search` table
**Severity**: Minor (cosmetic)
**PRD Reference**: §7.1 (tag = relative dir path; no charset restriction on
directory names) and §6.1 (`--list`/`--search` table).
**Expected Behavior**: Table columns align regardless of the tag's character set.
**Actual Behavior**: `ui.padRight` sizes columns by byte length (`len(s)`), so a
multi-byte tag like `café` (4 display columns, 5 UTF-8 bytes) is under-padded and
shifts its row's NAME/DESCRIPTION columns left. The code comment in `ui.go`
asserts "tags are relative dir paths of the same [ASCII]" — that assumption is
wrong: only the `name` field is restricted to `a-z0-9-`; directory names (and thus
tags) are unrestricted.
**Steps to Reproduce**:
```bash
mkdir -p /tmp/u/skills/café /tmp/u/skills/ascii
printf -- '---\nname: caf\ndescription: café skill\n---\n\n# x\n' > /tmp/u/skills/café/SKILL.md
printf -- '---\nname: ascii\ndescription: ascii skill\n---\n\n# x\n' > /tmp/u/skills/ascii/SKILL.md
SKPP_SKILLS_DIR=/tmp/u/skills ./skpp --list | cat -A
# The café row's DESCRIPTION column starts one column early vs the ascii row.
```
**Suggested Fix**: Either compute display width with `golang.org/x/text/width` /
`runewidth` (adds a dependency the PRD §4/§7.3 deliberately avoids), or document
that tables are byte-aligned and best-effort for non-ASCII tags. Lowest-risk: leave
behavior as-is but fix the incorrect code comment so future maintainers don't rely
on the false ASCII assumption.

### Issue 3: `.gitignore` deviates from the PRD §16 specification
**Severity**: Minor
**PRD Reference**: §16 (gives the exact `.gitignore` contents).
**Expected Behavior**: §16 specifies exactly:
```
/skpp
/dist
*.test
*.out
.DS_Store
```
**Actual Behavior**: The shipped `.gitignore` adds `/build`, `.env`, `.env.*`, and
`.pi-subagents/`. The extras are reasonable hygiene and do not cause the example
skill to be skipped (it is committed), but they are a deviation from the spec's
literal contents.
**Steps to Reproduce**: `cat .gitignore` and diff against §16.
**Suggested Fix**: Either trim to the §16 set, or update the PRD to bless the extra
entries. Low priority; flagging for spec/impl alignment.

### Issue 4: `--search` does not match `metadata.aliases` or `metadata.category`
**Severity**: Minor (spec inconsistency / UX)
**PRD Reference**: §6.1 (authoritative search field list: tag, name, description,
`metadata.keywords`) vs §10 (states `aliases`/`category`/`keywords` "exist only to
enrich `skpp --search` and tag aliases").
**Expected Behavior**: Ambiguous in the PRD. §10 implies aliases enrich `--search`;
§6.1's explicit field list excludes them. The implementer correctly followed the
more specific §6.1.
**Actual Behavior**: `skpp --search <alias>` returns "no skills matched" (exit 1),
even though `skpp <alias>` (resolution by alias, §7.2 step 4) works. A user who
reads §10 and adds aliases may expect search to find them.
**Steps to Reproduce**:
```bash
mkdir -p /tmp/a/skills/aliased
printf -- '---\nname: aliased\ndescription: x\nmetadata:\n  aliases: [my-alias]\n---\n\n# x\n' > /tmp/a/skills/aliased/SKILL.md
SKPP_SKILLS_DIR=/tmp/a/skills ./skpp --search my-alias   # no match (exit 1)
SKPP_SKILLS_DIR=/tmp/a/skills ./skpp my-alias            # resolves fine
```
**Suggested Fix**: Decide which spec wins and make §6.1 and §10 agree. Either (a)
extend `search.matches` to also scan `Aliases` (and optionally `Category`), or (b)
tighten §10's wording so it no longer implies aliases/category drive `--search`.
The code change for (a) is one extra field check in `internal/search/search.go`.

### Issue 5: Combined short flags and `--flag=value` are rejected
**Severity**: Minor (UX / POSIX convention)
**PRD Reference**: §6 ("Flags use POSIX double-dash long form + single-dash short
forms").
**Expected Behavior**: Debatable. Common CLIs accept `-vh` (combined shorts) and
`--version=x` (`=` value syntax). The PRD does not mandate either.
**Actual Behavior**: Both forms are treated as unknown flags and exit 2:
```bash
./skpp -vh            # unknown flag '-vh', exit 2
./skpp --version=x    # unknown flag '--version=x', exit 2
```
`parseArgs` matches exact token strings (`case "--version", "-v"`), so any
combined or `=`-bearing token falls through to the unknown-flag branch.
**Suggested Fix**: Optional. If desired, split combined shorts and strip `=value`
in `parseArgs` before the switch. Low priority; the individual flags all work.

### Issue 6: Combining two listing modes silently picks one by dispatch order
**Severity**: Minor (UX)
**PRD Reference**: §6.3 (mutual exclusivity is scoped to tag+mode and check+X).
**Expected Behavior**: Debatable. §6.3 does not forbid mode+mode combos.
**Actual Behavior**: `skpp --list --search foo`, `skpp --all --list`, and
`skpp --path --list` do not error; the first-matched mode in `run()`'s dispatch
order wins (`--path` > `--list` > `--search` > `check` > `--all`). A user typing
`--all --list` expecting an error (or a specific mode) gets `--list` silently.
**Steps to Reproduce**:
```bash
./skpp --list --search foo   # runs --list (shows all), ignoring --search
./skpp --all --list          # runs --list, ignoring --all
```
**Suggested Fix**: Optional. Either extend `exclusivityError` to reject any
combination of two+ listing modes, or document the dispatch precedence. The
current behavior is internally consistent and §6.3-compliant, so this is polish.

### Issue 7: `check` cannot report "skill dir has no SKILL.md"
**Severity**: Minor (effectively moot)
**PRD Reference**: §9 ("ERROR: skill dir has no SKILL.md").
**Expected Behavior**: §9 lists this as an ERROR condition.
**Actual Behavior**: This rule is unimplemented and cannot ever fire, because
`discover.Index` defines a skill as "any directory that directly contains a
SKILL.md" — it only emits directories that have one. A grouping directory with no
SKILL.md is simply not a skill and is never inspected. The code documents this as
deliberate (a heuristic would false-positive on legitimate grouping dirs).
**Suggested Fix**: None required for correctness. If §9 fidelity matters, either
remove the rule from the PRD or define what "looks like a skill but lacks
SKILL.md" means (e.g. a dir containing `scripts/` or `references/` but no
SKILL.md) and scan for it. Low priority.

## Testing Summary

- **Total tests performed**: ~70 distinct scenarios across happy path, edge
  cases, error handling, adversarial input, permissions, unicode, and
  integration.
- **Passing**: All §13 acceptance criteria; all core contracts (resolution,
  precedence, atomicity, exit codes, stdout discipline, discovery, validation,
  color/TTY gating, completions syntax, install.sh, pi integration).
- **Failing**: 1 Major (Issue 1), 6 Minor (Issues 2–7).
- **Areas with good coverage**: tag resolution precedence (canonical/basename/
  name/alias, ambiguity short-circuit), discovery rules (env/sibling/walk-up +
  fall-through on bad env), error/atomicity contract, frontmatter parsing (BOM,
  CRLF, no-block, broken YAML, length limits, name charset), name/description
  validation, exit-code matrix, help/version precedence, symlink install.
- **Areas needing more attention**: surfacing the discovery source to the user
  (Issue 1); non-ASCII display width (Issue 2); spec consistency between §6.1
  and §10 for searchable fields (Issue 4).
