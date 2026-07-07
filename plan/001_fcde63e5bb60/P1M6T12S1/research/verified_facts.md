# Verified Facts — P1.M6.T12.S1 (`skills/example/SKILL.md`, PRD §11)

> The ONE shipped example skill. This is the same deliverable previously tracked as
> P1.M6.T1.S1; its exhaustive research lives at
> `plan/001_fcde63e5bb60/P1M6T1S1/research/verified_facts.md` (read it first). This
> note records what is **authoritative NOW** (all upstream gates P1.M2–P1.M5 are
> landed and green) and supersedes only the parts the older note flagged as
> "pending". Verified against `/home/dustin/projects/skpp`, module
> `github.com/dabstractor/skpp`, go 1.25, yaml.v3 v3.0.1.

## §1. The exact file content (byte-verified, 569 bytes)

Write `skills/example/SKILL.md` with EXACTLY this content — the PRD §11 block
minus the four-backtick wrapper (the outer `````markdown` fence is rendering-only;
the inner ``` ```bash ``` fence is real file content). The PRD sentence
"No other skills ship in this repo." is PROSE after the closing fence, NOT a line
in the file:

```markdown
---
name: example
description: >
  Reference example skill for skpp. Demonstrates the required frontmatter and
  how skpp resolves a tag to an absolute path. Safe to delete once you add real skills.
metadata:
  keywords: [example, demo, skpp]
  category: meta
---

# Example Skill

This skill exists only so `skpp` has something to resolve.

Try:

```bash
skpp example                       # prints this directory's absolute path
skpp -f example                    # prints .../skills/example/SKILL.md
pi --skill "$(skpp example)"       # loads this skill into pi
```
```

## §2. The content is spec-compliant AND the discover package already accepts it

Parsed by the LANDED `discover.ParseFrontmatter` (`internal/discover/discover.go`,
yaml.v3) against a temp copy:

| field       | parsed value                              | §9 verdict |
|-------------|-------------------------------------------|------------|
| HasFM       | `true`                                    | OK         |
| name        | `example`, matches `^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`, len 7 ≤ 64 | OK |
| description | folded `>` → 162 chars (≤ 1024)           | OK (no WARN) |
| keywords    | `[example demo skpp]` (flow-seq → []string via `toStringSlice`) | OK |
| category    | `meta`                                    | OK         |
| aliases     | absent → `[]` (nil)                       | OK         |
| body        | present (`# Example Skill` + Try: block)  | OK         |

No `discover`/`resolve`/`ui`/`search`/`check` package changes are required. This
task touches NO `.go` file.

## §3. What changed since the older P1.M6.T1.S1 note — `skpp check` is NOW WIRED

The older note (§3) flagged `skpp check` as "NOT wired yet" because P1.M4.T10.S1
(`internal/check`) and the `case "check":` dispatch were pending. **Both are now
LANDED:**

- `main.go:184` → `case "check":` dispatch exists.
- `internal/check/check.go` implements PRD §9 (`OK`/`WARN`/`ERROR` report, exit 1
  on any ERROR).
- `go build -o /tmp/skpp-verify .` → exit 0 (verified today).
- `go test ./...` baseline is green (212 `Test*` funcs across the module).

So the PRD §13 acceptance line `./skpp check` is a FIRST-CLASS gate now, not a
conditional one. The check-independent probe at
`plan/001_fcde63e5bb60/P1M6T1S1/research/validate_example_probe.go` (`//go:build
ignore`) remains available as a belt-and-braces frontmatter validator that does
not depend on `internal/check`, but it is OPTIONAL.

## §4. `skills/` does not exist yet — created fresh by writing this one file

`git ls-files` shows no `skills/` directory and no `.gitkeep` anywhere. There is
nothing to remove. Writing `skills/example/SKILL.md` (the write tool auto-creates
parent dirs) is the entire creation step.

## §5. Location & the "not a pi auto-discovery location" guardrail (PRD §2.2)

- Sibling-of-binary rule (§8.2) resolves the store to `<repo>/skills`, i.e.
  `/home/dustin/projects/skpp/skills`. The example lives at
  `/home/dustin/projects/skpp/skills/example/SKILL.md`.
- `~/.pi/agent/skills` is a real pi auto-discovery location (contains
  `agent-browser`, `mdsel`, …). `<repo>/skills` is a DIFFERENT path. Satisfies
  PRD §2 constraint 2: skills load ONLY via `pi --skill "$(skpp <tag>)"`, never
  auto-discovered.
- Assertable: `test "$(./skpp --path)" = "$PWD/skills"`; printed path is NOT
  under `$HOME/.pi`.

## §6. Validation commands verified working TODAY

| command                                  | expected                                     | status |
|------------------------------------------|----------------------------------------------|--------|
| `go build -o skpp .`                     | exit 0                                       | ✅ |
| `go build ./...`                         | exit 0 (probe present, build-ignored)        | ✅ |
| `go test ./...`                          | green (no .go touched)                       | ✅ |
| `./skpp example`                         | `$PWD/skills/example`, exit 0                | ✅ (M3.T8.S1 landed) |
| `./skpp -f example`                      | `$PWD/skills/example/SKILL.md`, exit 0       | ✅ (M3.T8.S2 landed) |
| `./skpp --list`                          | `example` row (TAG/NAME/DESCRIPTION), exit 0 | ✅ (M2.T6.S1 landed) |
| `./skpp --path`                          | `$PWD/skills`, exit 0                        | ✅ (M1.T3.S1 landed) |
| `./skpp check`                           | `OK   example (example)`, exit 0             | ✅ NOW wired (M4.T10 + M5.T11) |
| `./skpp --search skpp`                   | `example` row (matches keyword `skpp`), exit 0 | ✅ (M4.T9 landed) |

## §7. Hard scope guardrail — EXACTLY one skill (PRD §2 constraint 4, §17)

This task creates ONLY `skills/example/SKILL.md`. Do not add a second skill, do
not add `scripts/`/`references/`/`assets/` siblings (the PRD §11 example has
none), do not add a manifest/index file (PRD §2 constraint 1, §17). The PRD's
"No other skills ship in this repo." is a constraint restatement, not file
content.
