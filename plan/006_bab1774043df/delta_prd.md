# Delta PRD — `--link` multi-target (batch linking)

> **Delta from:** session 005 (`plan/005_d9b30e368811/`), which is **Complete and green**.
> **Scope of THIS session:** implement PRD §8.4 / Decision 23 (new) + the §6.1/§6.3/§6.4/§14 adjustments — make `--link` accept **one or more** directories in a single invocation. Do not change any other contract.
> **Size:** ONE medium feature. Single phase, single milestone, five tasks. The repo is otherwise complete and green — this builds on the finished single-target `--link`.

---

## 1. What changed (diff summary)

The ONLY delta is **`--link` batch (multi-target) linking**. Verified by `diff plan/005.../prd_snapshot.md PRD.md` — every changed hunk is about `--link`:

| PRD location | Change |
|---|---|
| §6.1 table row | `--link <dir>` → **`--link <dir> [<dir>...]`**; stdout = one link path per success (input order); exit `0` if all link / `1` if any fail / `2` if no dir follows |
| §6.3 | `--link` is now the sole mode that **collects trailing positionals** (once seen, every following non-flag token is a link dir, never a tag); pass `--link` at most once |
| §6.4 | per-directory independent validation; mixed stdout/stderr; non-atomic partial success; `--link` with **no** following dir at all → `skilldozer: --link requires at least one path to a skill directory`, exit `2` |
| §8.4 | title `--link <dir> [<dir>...]`; new "Batching (headline behavior)" block; restructured numbered steps; new "Exit code and partial success" step; new multi-dir examples |
| §13 | new acceptance asserts: `multi-link OK` (both link, input order), `multi-link-partial OK` (mixed batch, exit 1, bad dir named on stderr), `link-non-skill-refused OK` (degenerate single-bad-dir batch) |
| §14.1 / §14.2 / §14.6 | `--link <tab>` → dirs (first); **`--link d1 <tab>` → dirs (every position after `--link` completes dirs)** |
| Decision 21 | updated to reference multi-target (decision 23) |
| Decision 23 | **NEW** — batch linking semantics: one-or-more dirs, non-atomic link-as-you-go, exit 1 if any fail (successful links remain), zero dirs ⇒ exit 2 |

**Nothing else changed.** No other flag, no discovery, no config, no completion-listing (§14.7, done in session 005). The §14.4 lockstep is NOT triggered (no flag add/rename/remove — `--link` already exists); only its *behavior* for the `--link` slot changes.

---

## 2. Goal of this delta

`skilldozer --link` currently links exactly **one** external skill directory per invocation (PRD §8.4 as implemented in main.go `runLink`, session 004). Linking several projects is a common setup step, and repeating `--link` per directory is needless friction. This delta extends `--link` to **batch**: one `--link`, then one-or-more directory arguments, linked in input order with non-atomic partial success.

```bash
# before (still works — degenerate n=1 batch)
skilldozer --link ~/projects/agent-browser

# after (new)
skilldozer --link ~/projects/agent-browser ~/projects/agent-builder ~/projects/mdsel
# stdout: one absolute link path per success, in input order
```

---

## 3. Behavioral contract (the part that must be exactly right)

### 3.1 Parsing model change (PRD §6.3 / §8.4 step 2)

`--link` switches from a **value-taking flag** (captures the single next token as `c.linkTarget`) to a **positional-collecting mode** (once `--link` is parsed, every following non-flag token is appended to the link-target list, never to `c.tags`).

- `skilldozer --link a b c` → link dirs `a`, `b`, `c` (in order).
- `skilldozer --link=<dir> b c` → the `=`-form supplies the **first** directory; `b` and `c` are further positionals that add to the list (PRD §8.4 step 2).
- A dashed token after `--link` is **not** a link dir (only non-flag positionals are collected). It is parsed as a flag in the usual way — a mode flag trips mutual exclusivity (exit 2); `--link` a second time is "pass `--link` at most once" (error). Edge cases where no dir is collected are handled by §3.3.
- The bare-positional namespace for tags stays reserved: positionals are only collected as link dirs **because `--link` was seen first**. Without `--link`, every positional is a tag (unchanged).

### 3.2 Linking (PRD §8.4 step 3 + "Exit code and partial success")

For each directory, **in input order**, run the existing absolutize → validate → conflict → symlink sequence (current `runLink` body, §8.4 steps 3a–3e). Per directory:

- success ⇒ print the link path to **stdout** (one per line) + the `Linked <path> -> <target> (found via <rule>)` line to **stderr**;
- failure (not a dir / inside store / no `SKILL.md` / non-symlink collision / symlink/IO error) ⇒ print **one stderr line naming that directory** and continue to the next.

Processing is **non-atomic (link-as-you-go)**: a valid dir links before the next is checked, so a batch may leave some symlinks created and others not.

### 3.3 Exit codes (PRD §6.1 row / §6.4)

- `0` iff **every** directory linked successfully;
- `1` iff **any** directory failed (successful links remain in place — linking is idempotent, so re-running `--link <only-the-failed-dir>` is the documented fix);
- `2` iff `--link` was seen but **zero** directories were collected (no `=value` and no following positional) → stderr `skilldozer: --link requires at least one path to a skill directory`, nothing on stdout.

The single-directory case is the degenerate n=1 batch and **must preserve** the existing guarantee: one bad dir ⇒ nothing on stdout, exit `1`.

### 3.4 Completions (PRD §14.1 / §14.2 / §14.6)

Every position **after** `--link` completes **directories** (not just the first): `skilldozer --link <tab>` and `skilldozer --link d1 <tab>` (and `d1 d2 <tab>`, …) all offer file/dir completion. This is a behavior change to the `--link` slot in all three completion files.

---

## 4. Implementation map (what exists today — DO NOT re-implement from scratch)

Build on the finished single-target `--link`. Verified line numbers against current `main.go` / `main_test.go`:

| Symbol / location | Current state | What this delta does |
|---|---|---|
| `config.link bool` / `config.linkTarget string` / `config.linkMissingValue bool` (main.go:176–178) | single target | **change parsing model** — collect a slice; see T1 |
| parseArgs `--link` `=`-form case (main.go:266–273) | sets `c.linkTarget = val`; empty ⇒ `linkMissingValue` | seed first target (or zero-dir marker); subsequent positionals collected |
| parseArgs `--link` next-token case (main.go:403–429) | captures `args[i+1]` as `linkTarget`; dashed/last-token ⇒ `linkMissingValue` | enter "link-collect mode"; trailing non-flag tokens append to the list |
| `run()` `linkMissingValue` → exit 2 block (main.go:604–610) | prints `skilldozer: --link requires a path to a skill directory` | message → `... requires at least one path ...`; the zero-collected-dirs check now runs **after full parse** |
| `exclusivityError` link block (main.go:941–953) | rejects `'--link' cannot be combined with tag arguments` | **remove the `hasTags` branch for link** — positionals after `--link` are link dirs, not tags; keep mode-mutex |
| `runLink(c, …)` (main.go:1424–1487) | single target; returns 0/1 | **loop** over targets; per-target stdout/stderr; exit 0 if all / 1 if any |
| dispatch `if c.link { return runLink(...) }` (main.go:651) | unchanged call | unchanged (signature may take the slice) |
| completions/skilldozer.bash (~line 43) | `--store\|--init\|--link) COMPREPLY=($(compgen -d …))` keyed on `prev` | detect `--link` anywhere in `words`; offer dirs for every trailing positional |
| completions/_skilldozer (zsh, ~line 51) | `'--link[...]:directory:_files'` (consumes one value) | every position after `--link` → `_files` (state-based) |
| completions/skilldozer.fish (~line 39) | `complete -l link … -r` (one value) | conditional: when `--link` seen, every positional offers dirs |
| main_test.go:3489–3545 | `TestParseArgsLinkNextToken/Equals/NoValue/EqualsEmpty` assert single `c.linkTarget` | **update** for the new slice model (T3) |
| main_test.go:~3575 `TestRunLinkWithTagExits2` | asserts `--link /tmp/foo sometag` ⇒ exit 2 | **MUST CHANGE** — `sometag` is now a 2nd link target, not a forbidden tag (T3) |
| main_test.go runLink handler tests (~3599+) | single-target success/refresh/refuse/unconfigured | stay valid (degenerate n=1); add batch/partial/zero tests (T3) |
| README.md (~127–128, 158–175) | single-dir `--link` usage + "Linking skills" section | document batching + partial-success exit semantics (T5) |

**Reuse, don't rewrite:** the per-directory validate/conflict/symlink logic in the current `runLink` body (steps 2–7, main.go:1430–1487) is correct and unchanged — it becomes the loop body. `expandHome`, `filepath.Abs`, `skillsdir.HasSkillMD`, the non-symlink refusal, and the symlink-refresh behavior are all kept verbatim.

---

## 5. Documentation impact

- **Mode A — doc-with-work:** None beyond the README. No JSDoc/config-doc surfaces; the Go doc comments on `runLink` / the `--link` parse cases are updated inline as part of the code tasks (T1/T2). The completion-file lockstep comments (§14.4) are updated inline as part of T4.
- **Mode B — changeset-level docs:** YES → the README "Linking skills from elsewhere (`--link`)" section must reflect batching + partial-success exit codes (Task T5, runs last, depends on the implementing tasks). This is the only cross-cutting doc.

---

## 6. Out of scope (hard guardrails)

- ❌ No flag add/rename/remove — the §14.4 lockstep flag set is UNCHANGED; only the `--link` slot's *behavior* changes (T4 edits the slot, not the flag list).
- ❌ Do not touch `--init`, `--store`, `--search`, `--shell`, tag resolution, discovery, config, `skillsdir`, `ui`, `runCheck`, `runInit`, or the §14.7 listing machinery (done in session 005).
- ❌ Do not change `--link`'s per-directory validation rules (existing dir, not store/inside-store, ≥1 `SKILL.md`, non-symlink refusal, symlink refresh) — only the *loop* around them is new.
- ❌ Do not make batch linking atomic (no all-or-nothing rollback). PRD Decision 23 explicitly mandates non-atomic link-as-you-go.
- ❌ Do not change the single-directory guarantee: n=1 with a bad dir ⇒ nothing on stdout, exit `1`.

---

## 7. Acceptance (mirrors the new §13 asserts)

All existing acceptance still passes; the following NEW asserts must pass from a clean build:

```bash
# multi-link: one --link, then several directories (§8.4)
rm -rf /tmp/sd-link/store && mkdir -p /tmp/sd-link/store /tmp/sd-link/src/other /tmp/sd-link/notaskill
printf -- '---\nname: other\ndescription: Another linked skill.\n---\n# body\n' > /tmp/sd-link/src/other/SKILL.md
out=$(SKILLDOZER_SKILLS_DIR=/tmp/sd-link/store ./skilldozer --link /tmp/sd-link/src/linked /tmp/sd-link/src/other)
printf '%s\n' "$out" | grep -qx '/tmp/sd-link/store/linked' && printf '%s\n' "$out" | grep -qx '/tmp/sd-link/store/other' && echo "multi-link OK"   # both link paths, input order
test -L /tmp/sd-link/store/linked && test -L /tmp/sd-link/store/other                            # both symlinks created
# mixed batch: two valid + one invalid → valid ones link, exit 1, the bad dir named on stderr
out=$(SKILLDOZER_SKILLS_DIR=/tmp/sd-link/store ./skilldozer --link /tmp/sd-link/src/linked /tmp/sd-link/src/other /tmp/sd-link/notaskill 2>err); rc=$?
[ "$rc" = "1" ] && printf '%s\n' "$out" | grep -qx '/tmp/sd-link/store/linked' && grep -q 'notaskill' err && echo "multi-link-partial OK"
# single bad dir (degenerate batch): nothing on stdout, exit 1
out=$(SKILLDOZER_SKILLS_DIR=/tmp/sd-link/store ./skilldozer --link /tmp/sd-link/notaskill 2>/dev/null); rc=$?
[ -z "$out" ] && [ "$rc" = "1" ] && echo "link-non-skill-refused OK"
./skilldozer --link >/dev/null 2>&1; [ "$?" = "2" ] && echo "link-missing-value OK"               # no dir at all → exit 2
```

Plus: `go build ./...` clean, `go test ./...` 100% green (existing single-target link tests updated, not deleted), and the n=1 happy path (`skilldozer --link <one-good-dir>` ⇒ one path on stdout, exit 0) still works.

---

## 8. Backlog

### Phase P1 — Multi-target `--link` (batch linking)

Implements PRD §8.4 / Decision 23 + the §6.1/§6.3/§6.4/§14 adjustments. One cohesive milestone; the parsing model (T1), the loop (T2), and exclusivity are tightly coupled and land together before tests/completions/docs.

#### Milestone P1.M1 — Collect, batch-link, complete, and document multi-target `--link`

##### Task P1.M1.T1 — parseArgs + config + exclusivity: `--link` collects one-or-more positionals

Convert `--link` from a value-taking flag to a positional-collecting mode. This is the parsing-model foundation; T2 (the loop) depends on it.

**Subtask P1.M1.T1.S1** — config struct + parseArgs collection model
- `config` (main.go:176–178): replace the single-target fields with a collection model. Keep `link bool` (the mode sentinel). Replace `linkTarget string` with a slice (e.g. `linkTargets []string`). Replace `linkMissingValue bool` with a "zero dirs collected" signal — simplest is to derive it at run() time from `c.link && len(c.linkTargets)==0` (the collection is only complete after the full parse), so the explicit boolean may be dropped OR retained as a parse-time marker for the empty `=`-form; pick whichever keeps the run() exit-2 check cleanest and document it.
- parseArgs `=`-form case (main.go:266–273): `--link=<dir>` — non-empty value seeds the **first** target (append to the slice); empty value records the no-value/zero-dir signal. Set `c.link = true`.
- parseArgs next-token case (main.go:403–429): change from "consume `args[i+1]` as the single target" to "enter link-collect mode". Once `--link` is seen, every subsequent **non-flag** positional token (in the default branch, main.go ~437+) is appended to `c.linkTargets` instead of `c.tags`. A dashed token is NOT a link dir (it parses as a flag). The `--` end-of-options separator (endOfOpts) still applies: after `--`, every token is a positional; under link-collect mode those positionals are link dirs (not tags).
- "pass `--link` at most once": a second `--link` token is an error (exit 2). Implement in parseArgs (detect repeat) or exclusivity (T1.S2) — keep it simple and documented.
- Doc comments on the two `--link` parse cases + the `config` fields updated to describe collection, not single capture.
- DOCS (Mode A, inline): update the Go doc comments on the `--link` cases and `config.link*` fields as part of this edit.

**Subtask P1.M1.T1.S2** — exclusivityError: rework the tag-after-link rule
- `exclusivityError` link block (main.go:941–953): **remove** the `if hasTags { return "'--link' cannot be combined with tag arguments" }` branch — positionals after `--link` are now link dirs, not tags, so `hasTags` is never true when `c.link` is set (they were collected into `linkTargets` instead). Keep the mode-mutex branch (`--link` vs `--check/--init/--completions/--path/--list/--search/--all` → exit 2) and fold the "pass `--link` at most once" check in here if not done in parseArgs.
- Verify the `--link <dir> --check` path still exits 2 (mode-mutex), and `--link <dir> <dir>` no longer exits 2 (it is a valid 2-target batch).
- run() exit-2 message (main.go:604–610): if keeping a pre-dispatch zero-dir check, change the text to `skilldozer: --link requires at least one path to a skill directory` (matches PRD §6.4). The check now fires on `c.link && len(c.linkTargets)==0` (post-parse), not on a parse-time next-token peek.

##### Task P1.M1.T2 — runLink: non-atomic batch loop, per-target output, exit 0/1

Depends on T1 (consumes `c.linkTargets`). Turn the single-target `runLink` body into a loop.

- `runLink(c, stdout, stderr)` (main.go:1424–1487): resolve the store ONCE up front (steps 1, unchanged) — an unconfigured store still ⇒ one-line fix to stderr, exit `1`, nothing on stdout (no per-dir output). Then **for each target in `c.linkTargets`, in input order**, run the existing per-directory sequence (absolutize → validate existing-dir → validate not-store/inside-store → validate HasSkillMD → basename/linkPath → conflict handling → os.Symlink) as the loop body.
- Per target: on success print the link path to **stdout** (one line) + `Linked <path> -> <target> (found via <rule>)` to **stderr**; on any validation/conflict/IO failure print **one stderr line naming that directory** and continue. Track a `failed bool`; do NOT short-circuit.
- Return `0` if `!failed` (all linked), `1` if any failed. The store-resolution failure (step 1) returns `1` before the loop with empty stdout, as today.
- The per-directory stderr messages should still name the offending directory (they already use `%q` on the target/linkPath). Keep the existing message wording; only the loop + return value are new.
- Degenerate n=1 batch must behave exactly like today's single-target handler (one good dir ⇒ one stdout path, exit 0; one bad dir ⇒ nothing on stdout, exit 1).
- Doc comment on `runLink` updated to describe the batch loop + non-atomic partial success (Decision 23).
- DOCS (Mode A, inline): the `runLink` doc comment is the developer-facing reference.

##### Task P1.M1.T3 — Tests: update single-target tests for the new model + add batch tests

Depends on T1 + T2. Mirror the existing `TestParseArgsLink*` / `TestRunLink*` patterns (main_test.go:3489+). Use the existing `mkExtSkill` helper.

- **Update** the parseArgs tests for the slice model: `TestParseArgsLinkNextToken` / `TestParseArgsLinkEquals` assert `c.linkTargets == []string{"/path/to/skill"}` (or equivalent) instead of `c.linkTarget`. `TestParseArgsLinkNoValue` / `TestParseArgsLinkEqualsEmpty` assert the zero-dir signal under the new model.
- **Add** parseArgs tests: `--link a b c` → `linkTargets == [a,b,c]`, no tags; `--link=<d> b c` → `linkTargets == [<d>,b,c]`; second `--link` ⇒ error; under link-collect mode, trailing positionals do NOT land in `c.tags`.
- **Update** `TestRunLinkWithTagExits2` (main_test.go:~3575): `--link /tmp/foo sometag` no longer exits 2 — `sometag` is a 2nd link target. Replace with a test that `--link <good-dir> <good-dir2>` links both and exits 0 (or repurpose to assert the n=2 batch).
- **Keep green** `TestRunLinkWithCheckExits2` (mode-mutex still exit 2) and all single-target `runLink` handler tests (they are the n=1 degenerate batch).
- **Add** runLink batch tests mirroring §7: (a) two good dirs ⇒ both link paths on stdout in input order, exit 0, both symlinks created; (b) two good + one bad (no SKILL.md) ⇒ good ones on stdout, bad dir named on stderr, exit 1, good symlinks remain; (c) single bad dir ⇒ nothing on stdout, exit 1 (degenerate); (d) `--link` with zero dirs ⇒ exit 2, empty stdout, message `requires at least one path`; (e) store-unconfigured in a batch ⇒ one-line fix, exit 1, empty stdout.
- Assert stream discipline: successes on stdout only, failures on stderr only.
- DOCS (Mode A): none — test-only.

##### Task P1.M1.T4 — Completions: every position after `--link` completes directories

Independent of T1–T3 (completion files only). Updates the `--link` slot in all three files (§14.4 lockstep flag list UNCHANGED — `--link` is already listed; only its slot behavior changes).

- **bash** (completions/skilldozer.bash, ~line 43): currently `prev` is `--store|--init|--link` ⇒ `compgen -d`. That only offers a dir for the FIRST token after `--link`. Add a check that scans `words` (the `_init_completion` array) for `--link`: if present AND the current token is a positional (not a value of `--store`/`--init`/`--search`/`--shell`), offer `compgen -d` for **every** trailing positional. Keep `--store`/`--init` one-value behavior intact. Update the inline §14.4 lockstep comment for the `--link` slot.
- **zsh** (completions/_skilldozer, ~line 51): `'--link[...]:directory:_files'` consumes one value. Use a `_arguments` state (`-C` / `->state`) or a `'*:: : _files'`-style trailing rule so that once `--link` is present, every remaining positional completes directories. Keep `--store`/`--init` one-value. Update the inline comment.
- **fish** (completions/skilldozer.fish, ~line 39): `complete -l link … -r` consumes one value. Add a conditional completion (`-n` condition checking that `--link` has been seen in the command line, e.g. via `__fish_seen_subcommand_from` is NOT applicable — use `string match -q -- '* --link *' (commandline)` or an equivalent helper) that offers directory completion (`-a "(__fish_complete_directories ...)"` or `-r`-style) for **every** positional after `--link`. Keep the skills-first default for the no-`--link` case. Update the inline comment.
- **No new flag in any file** — `--link` is already in all three; only the slot's completion target changes. Run `./skilldozer --completions --shell {bash,zsh,fish}` after rebuild and eyeball that `--link` is still advertised and the dir-completion follows it across positionals.
- DOCS (Mode A, inline): the §14.4 lockstep comments in each file note the `--link` slot now completes dirs at every position.

##### Task P1.M1.T5 — README: document `--link` batching + partial-success (Mode B)

Depends on T1–T4. Runs last. Updates README.md "Linking skills from elsewhere (`--link`)" (~lines 158–175) and the one-line usage (~lines 127–128).

- Change the usage line from `skilldozer --link <dir>` to `skilldozer --link <dir> [<dir>...]` and show a multi-dir example (`skilldozer --link ~/projects/agent-browser ~/projects/agent-builder ~/projects/mdsel`).
- In the "Linking skills" section: state that one `--link` links one-or-more dirs in input order; stdout is one absolute link path per success; exit `0` if all link, `1` if any fail (successful links remain — re-run `--link <failed-dir>` to finish); `--link` with no dir ⇒ exit `2`.
- Keep the existing single-dir narrative (refresh on re-run, refuses non-symlink, absolutizes `~`, resolves via the symlink) — it is the n=1 case and still true.
- Preserve all other README content (install, store-resolution, completions §14.7 disclosure from session 005, etc.).
- DOCS (Mode B): this IS the changeset-level documentation task for this delta; it depends on the implementing tasks and runs last.
