# Architectural Decisions — skpp Bug Fix 001

These decisions resolve the ambiguities the bug report leaves open. Each
is grounded in the codebase research and the PRD spec.

---

## D1 — Issue 1: Source label goes to stderr (not stdout, not a new flag)

**Decision**: Print `(found via <Source.String()>)\n` to **stderr**.

**Rationale**: The §13 acceptance gate `test "$(./skpp --path)" = "$PWD/skills"`
locks stdout to exactly `<dir>\n`. A new flag (e.g. `--verbose`) would not
help the typo'd-env-var scenario — the user runs plain `--path` and needs to
see the rule without knowing to ask for it. stderr is invisible to `$(...)`
but visible on a terminal, which is exactly where a human debugging discovery
is looking. This matches the bug report's suggested fix and requires no new
flag surface.

---

## D2 — Issue 2: Use stdlib utf8.RuneCountInString (NOT a third-party dep)

**Decision**: Add a `displayWidth(s string) int` helper using
`unicode/utf8.RuneCountInString`. Do NOT add `golang.org/x/text/width` or
`runewidth`.

**Rationale**: PRD §4/§7.3 deliberately keeps `gopkg.in/yaml.v3` as the ONLY
third-party dependency. Rune count fixes the common misalignment vectors
(é, —, smart quotes, most emoji that render 1-cell-wide) using stdlib only.
Wide CJK runes (display width 2) remain imperfect — this limitation is
documented in the code comment. This is strictly better than the current
byte-length approach with zero dependency cost.

---

## D3 — Issue 3: Trim .gitignore to PRD §16 spec (do NOT bless extras)

**Decision**: Remove `/build`, `.env`, `.env.*`, `.pi-subagents/`. Restore the
exact §16 set.

**Rationale**: PRD.md is read-only (human-owned). The spec is explicit about the
5-entry set. The extras are "reasonable hygiene" but represent undocumented
deviation. Bringing the file into spec compliance is the only action that
resolves the discrepancy without modifying PRD.md. If maintainers want the
extras, they update §16 themselves.

---

## D4 — Issue 4: Extend search to match Aliases AND Category (§10 wins over §6.1)

**Decision**: `matches()` scans Aliases and Category in addition to the current
four fields.

**Rationale**: §10 explicitly states that `metadata.keywords`,
`metadata.category`, and `metadata.aliases` "exist only to enrich `skpp --search`
and tag aliases." This is a direct, intentional statement that all three enrich
search. §6.1's field list omits them, but §6.1 is a summary table — §10's
prose is the more specific, intentional claim. A user who adds aliases (which
already drive tag resolution per §7.2 step 4) reasonably expects search to find
them. The existing test `TestSearchDoesNotMatchCategoryOrAliases` encodes the
wrong interpretation and must be inverted.

---

## D5 — Issue 5: Implement combined shorts + --flag=value normalization

**Decision**: Normalize tokens in `parseArgs` before the switch. Support:
- `--flag=value` for all long flags (value used by `--search`, ignored for bools)
- `-abc` combined short bool flags (expand to `-a -b -c`)
- `-sVALUE` attached short value (e.g. `-sfoo` → search "foo")

**Rationale**: POSIX convention. Low risk. The short flag set is small
(`v h p l a f s`). Value-taking is limited to `-s`/`--search`.

---

## D6 — Issue 6: Reject any 2+ listing-mode combination

**Decision**: Extend `exclusivityError` to treat `{path, list, searchMode, all}`
as mutually exclusive with each other (any 2+ → exit 2).

**Rationale**: Silent dispatch precedence is surprising. The user typing
`--all --list` expects either `--all` behavior or an error — not silent
`--list`. A loud error (exit 2) is consistent with the existing tags+mode and
check+mode exclusivity families. Also fix the stale dispatch-order comment.

---

## D7 — Issue 7: No code change; documentation-only

**Decision**: Do NOT implement a "no SKILL.md" heuristic. Confirm the existing
check.go comment is accurate. No action beyond verifying the comment.

**Rationale**: The bug report itself says "None required for correctness."
A heuristic (e.g. flagging dirs with `scripts/` but no `SKILL.md`) would
false-positive on legitimate grouping directories. The §9 rule is already
reframed in check.go:150 as "invalid SKILL.md frontmatter" (an unreadable/
malformed file is the actionable version of "no usable SKILL.md").
