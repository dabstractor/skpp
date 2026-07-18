# Verified Facts — P1.M3.T1.S1 (Mode B sweep: README overview + §13 acceptance for --link batch)

Mode B changeset-level documentation sweep for the --link multi-target batch feature (plan 006).
All facts read directly from source at PRP-write time. Repo: `/home/dustin/projects/skilldozer`.

## §0 — What this task IS (Mode B sweep) + the charter

- **IS**: the changeset-level doc sync (contract DOCS §5: "This IS the changeset-level documentation
  sync task (Mode B)"). It (a) RUNS the full §13 acceptance suite and (b) sweeps README's CROSS-CUTTING
  sections for stale single-target --link wording left by the Mode A per-subtask edits.
- P1.M1.T1.S1 (Go impl) is Complete and already updated README's MAIN --link section (Mode A) +
  the quick-start example. P1.M2.T1.S1 (completions, parallel/"Ready") updates the three completion
  FILES (multi-position dir completion) — disjoint from README. This task edits README's OVERVIEW
  sections only + verifies §13.

## §1 — The --link batch contract (PRD §8.4 / §6.1 — authoritative)

- `--link <dir> [<dir>...]` collects **one or more** directory positionals (batch). `--link` once;
  every following positional is a directory.
- `--link` is the SOLE mode that COLLECTS trailing positionals: once parsed, every following non-flag
  token is a directory to link (never a tag). `--link=<dir>` = first dir; further positionals add.
- Zero following dirs ⇒ `skilldozer: --link requires at least one path to a skill directory`, exit 2.
- Non-atomic partial success: each dir validated+linked independently in input order; exit 0 if all
  link, exit 1 if any fail (successful links remain). Mirrors ln/git add multi-target convention.
- Exclusive mode (mutually exclusive with all other modes, §6.3).
- PRD §17 line 120 + §6.1 line 138 + §8.4 line 149 all describe this BATCH behavior correctly.

## §2 — §13 acceptance: --link slice VERIFIED PASSING (run against built binary @ HEAD)

Ran the full --link block of PRD §13 against `go build -o skilldozer .`. Every case PASSED:
```
link OK                        (single: --link /tmp/sd-link/src/linked → /tmp/sd-link/store/linked)
resolve-linked OK              (linked resolves via the symlink)
link-refresh OK                (re-link refreshes the existing symlink)
multi-link OK                  (--link linked other → both paths, input order)
multi-link-partial OK          (linked other notaskill → exit 1, linked path on stdout, notaskill on stderr)
link-non-skill-refused OK      (single bad dir → nothing on stdout, exit 1)
link-missing-value OK          (--link with no dir → exit 2)
```
The changeset's Go side is sound. The implementer runs the FULL §13 suite (the non-link cases are
pre-existing and unaffected; the pi line needs `pi` installed — skip if absent, note it).

## §3 — README sections that are ALREADY batch-correct (Mode A landed them — NO edit)

- **Quick-start Usage block** (README:128-129): shows BOTH single and batch forms:
  `skilldozer --link ~/projects/agent-browser` and `skilldozer --link ~/projects/a ~/projects/b`.
- **Main "### Linking skills from elsewhere" section** (README:159-190): fully batch-correct —
  `<dir> [<dir>...]`, "Pass `--link` once; every positional after it is a directory to link",
  partial success, exit codes 0/1/2, refresh/refuse. (P1.M1.T1.S1's Mode A deliverable.)
- **Completions advertised flag list** (README:336): INCLUDES `--link` (15 long flags). ✓
- **Constraints section**: does not mention --link (no guardrail needed; the never-clobber refuse-
  non-symlink behavior is documented in the main --link section). ✓

## §4 — README sections with DRIFT (the sweep's edits)

### Edit 1 (REQUIRED) — Error contract / mode-flags discussion (README:152-155)
CURRENT (groups --link with single-value flags — misleading):
```
the whole store). `--link` is another such mode. The flags that take a value —
`--store`, `--search`, `--shell`, and `--link` — all exit 2 when given as the
last token with nothing after them, rather than guessing a value.
```
DRIFT: `--link` does NOT "take a value" (singular) — it COLLECTS one-or-more directory positionals
(a batch). The wording "guessing a value" doesn't fit --link. Contract (b): "should mention that
--link collects multiple positionals."
TARGET (pull --link out; keep exit-2-on-missing for all):
```
the whole store). `--link` is another exclusive mode: it collects **one or more**
directory positionals, so `--link` with nothing after it exits 2 rather than
linking nothing. The single-value flags — `--store`, `--search`, and `--shell` —
likewise exit 2 when given as the last token with nothing after them, rather than
guessing a value.
```

### Edit 2 (REQUIRED) — Completions value-completion enumeration (README:340)
CURRENT (groups --link with --init/--store "single directory"):
```
- `skilldozer --init <tab>`, `skilldozer --link <tab>`, and `skilldozer --store <tab>`
  offer directories (a path value); `skilldozer --search <tab>` offers nothing
  (free-text); `skilldozer --shell <tab>` offers the three supported shells —
  `bash`, `zsh`, and `fish`.
```
DRIFT: after P1.M2.T1.S1, `--link` offers directories at EVERY positional (`--link d1 <tab>` → dirs,
`--link d1 d2 <tab>` → dirs, …), UNLIKE --init/--store (single value). The current line conveys a
single value. Contract (b): "should note `skilldozer --link d1 <tab>` completes directories."
TARGET (pull --link out; note multi-position):
```
- `skilldozer --init <tab>` and `skilldozer --store <tab>` offer a single directory
  (a path value); `skilldozer --search <tab>` offers nothing (free-text);
  `skilldozer --shell <tab>` offers the three supported shells — `bash`, `zsh`, and
  `fish`. `skilldozer --link <tab>` offers directories too, and keeps offering them
  at **every** following positional (`--link d1 <tab>`, `--link d1 d2 <tab>`, …),
  because `--link` batches one or more directories.
```
NOTE: this edit documents the behavior P1.M2.T1.S1 lands (multi-position dir completion). P1.M3
depends on P1.M2, so at execution time the completion behavior is in place.

## §5 — Sweep checks (c) and (d) — NO FURTHER DRIFT FOUND

- **(c) stale single-target wording:** `grep -n -- '--link' README.md` shows every --link reference
  is batch-aware EXCEPT the two Edit spots (Error contract + completions enumeration). The main
  section, quick-start, and advertised flag list are all correct. Fixing the 2 edits satisfies (c).
- **(d) constraints/guardrails:** the README Constraints section does not mention --link and needs no
  mention (the never-clobber refuse-non-symlink behavior is in the main --link section; PRD §17 line
  263 references the "--link never clobber spirit" which the README main section already reflects).
  No edit. Document as verified-consistent.

## §6 — README voice rules (verified)

- §-PRD citations are SPARSE: `grep -no '§[0-9.]*' README.md` → 3 cites (§8.4 ×2 at 127/174, §6.1 at
  222). The EDITED sections (Error contract, Completions enumeration) use ZERO § cites. → keep new
  prose citation-free (plain words), matching local voice. Bold (`**one or more**`/`**every**`) is
  the house emphasis style (used elsewhere in the README).
- Inline `` `code` `` for commands/flags; em dashes; plain declarative prose; one idea per sentence.

## §7 — Parallel-execution / file-conflict consideration

- P1.M2.T1.S1 (completions, "Ready"/implementing) edits the THREE completion FILES only — NOT README.
  This task edits README.md ONLY. DISJOINT files → no merge conflict; land in either order.
- P1.M1.T1.S1 (Complete) already edited README's MAIN --link section (~line 159) + quick-start
  (~line 128). This task's edits are in DIFFERENT regions (Error contract ~line 152, Completions
  ~line 340) — disjoint from P1.M1.T1.S1's regions. No overlap.
- Edit 2 (completions section) documents P1.M2.T1.S1's landed behavior — at execution time (P1.M3
  after P1.M2) the multi-position dir completion is in place. Pin seams by ANCHOR TEXT (§4), not
  line number (the README may shift as siblings land).

## §8 — Scope boundary / what NOT to touch

- main.go / main_test.go / internal/* / completions/* / install.sh — UNCHANGED (all --link code +
  completions landed; this is doc-only). `go build/vet/test ./...` stays green (the doc-sanity check).
- PRD.md, tasks.json, prd_snapshot.md, .gitignore — READ-ONLY.
- README's main --link section (~line 159) + quick-start (~line 128) — P1.M1.T1.S1's Mode A territory;
  already correct, do NOT re-edit.

## §9 — Baseline (verified)

- `go build ./...` clean; `go test ./...` green.
- §13 --link slice: all 7 cases PASS (§2).
- README §-cite count = 3; --link appears batch-correct everywhere except the 2 Edit spots.
