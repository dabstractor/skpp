# Delta PRD — Subcommands → flags + skills-first completions

> **Status:** Ready for implementation. This delta modifies the existing, fully-working `skilldozer` CLI (all tests pass at HEAD `f30d5c5`).
> **Delta from:** session 003 (which implemented the `completion` subcommand + no-args implicit help).
> **Scope:** Convert the three bare-word reserved subcommands (`check`, `init`, `completion`) into long-only `--flags` (`--check`, `--init`, `--completions`) so the bare positional namespace belongs entirely to skill tags, then rewrite the three shell completion files to be skills-first and long-form-only. **Do not redesign discovery, resolve, check, config, or the embed mechanism** — they are untouched.

---

## 1. What actually changed (the delta)

A focused diff of `plan/003.../prd_snapshot.md` vs `plan/004.../prd_snapshot.md` shows **two coupled decisions**, not a feature explosion:

### Decision 19 (NEW) — Subcommand → flag (namespace safety)
- **OLD (what session 003 shipped):** `check`, `init`, `completion` are **bare-word reserved subcommands**. `skilldozer check` runs validation; a skill literally tagged `check` is shadowed.
- **NEW:** **No bare-word subcommands.** Every non-tag action is a `--flag`. The bare positional namespace is reserved for skill tags, so:
  - `skilldozer --check` runs validation; `skilldozer check` resolves the **tag** `check`.
  - `skilldozer --init [<dir>]` runs first-run setup; `skilldozer init` resolves the **tag** `init`.
  - `skilldozer --completions [--shell <name>]` emits the completion script (`completion` also **renamed singular → plural**); `skilldozer completions` resolves the **tag** `completions`.

### Decision 20 (NEW) — Completions are skills-first & long-form-only
- **OLD:** completions offered `check`/`init`/`completion` as exclusive first-positional subcommands, and offered both `--long` and `-short` flag forms.
- **NEW:**
  - A bare `skilldozer <tab>` shows **skills** — never the help menu, never a command list.
  - `skilldozer -<tab>` / `--<tab>` offers **long-form flags only** (`--check`, `--completions`, …); short aliases (`-a`, `-l`, `-s`, `-f`, `-p`, `-h`, `-v`) stay valid at runtime but are **not advertised**.
  - Prefix filtering (`a<tab>`, `--c<tab>`) is the shell's native job — the script offers the full list and lets the shell narrow it.

### Everything else in the diff is a consequence
The remaining ~40 diff hunks are mechanical renames flowing from decisions 19/20: every `skilldozer check` → `skilldozer --check`, `init` → `--init`, `completion` → `--completions` reference in §7.3, §8, §8.2, §8.3, §9, §12, §13 acceptance, §15 README outline, §18 build order, §19 decisions, plus the two new §17 guardrails and the rewritten §14.1/§14.2/§14.6.

### What is NOT in scope (do not touch)
- `internal/discover`, `internal/resolve`, `internal/check`, `internal/config`, `internal/search`, `internal/ui`, `internal/skillsdir` **logic** — unchanged. (Only the `skillsdir.ErrNotFound` *message string* changes: `init` → `--init`.)
- The `//go:embed` mechanism, `completionScript()`, `detectShell()`, `runCompletion()` dispatch — unchanged (the previous session's `plan/003_3ace946c2a5c/architecture/external_deps.md` research on embed is still valid and is reused).
- Short-form flags (`-v`/`-h`/`-p`/`-l`/`-a`/`-f`/`-s`) — **still valid at runtime** (§6.1); only the completion menu stops advertising them.
- The example skill `skills/example/SKILL.md` — uses only `skilldozer example` / `-f example` / `pi --skill`, so it needs no change.

---

## 2. Acceptance — the delta must make these true

Run from a clean `go build`. (These are the §13 assertions that *changed*; the unchanged ones still pass.)

```bash
go build -o skilldozer . && echo BUILD_OK

# Decision 19: subcommands are now FLAGS. Bare words resolve as tags (or fail as unknown tags).
./skilldozer --check                       # validation runs, exit 0 on clean store
./skilldozer --completions --shell bash 2>/dev/null | grep -q '_skilldozer_completion' && echo COMPLETIONS_BASH_OK
./skilldozer --completions --shell fish 2>/dev/null | grep -q 'complete -c skilldozer' && echo COMPLETIONS_FISH_OK
./skilldozer --init --store /tmp/sd-delta-store >/dev/null 2>&1 && test -d /tmp/sd-delta-store && echo INIT_FLAG_OK
# unsupported --shell ⇒ exit 2; undetectable shell ⇒ exit 1 (unchanged, now via --completions)
./skilldozer --completions --shell tcsh >/dev/null 2>&1; [ "$?" = "2" ] && echo BAD_SHELL_EXIT2_OK
out=$(env -u SHELL -u SKILLDOZER_SHELL ./skilldozer --completions 2>/dev/null); [ -z "$out" ] && echo COMPLETIONS_NO_SHELL_STDOUT_EMPTY

# Namespace safety: a bare word that used to be a reserved subcommand is now a TAG (or unknown).
out=$(./skilldozer check 2>/dev/null); rc=$?      # 'check' is not a shipped skill ⇒ unknown tag
[ -z "$out" ] && [ "$rc" = "1" ] && echo BARE_CHECK_NOW_TAG_OK

# Error message strings now reference --init (not bare init)
./skilldozer x 2>&1 | grep -q 'run `skilldozer --init`' && echo UNCONFIGURED_HINT_OK

# Completions embed the NEW files (rebuild required per §14.6 lockstep)
./skilldozer --completions --shell bash 2>/dev/null | grep -q '\-\-completions' && echo EMBED_HAS_COMPLETIONS_FLAG_OK
./skilldozer --completions --shell bash 2>/dev/null | grep -q '\-\-check' && echo EMBED_HAS_CHECK_FLAG_OK
# long-form-only: short flags must NOT appear in the embedded bash flag offer
! ./skilldozer --completions --shell bash 2>/dev/null | grep -Eq '\-\-version[ ]+-v' && echo LONG_FORM_ONLY_BASH_OK
```

Plus the **full existing suite must still pass** (`go test ./...`). No existing green test may regress except where it asserted the old bare-subcommand contract (those flip — see P1.M1.T2).

---

## 3. Documentation impact

- **Mode A (rides with the work):**
  - P1.M1: `main.go` doc comments (config struct fields, parseArgs cases, exclusivityError), the `usageText` constant (the user-facing `--help` surface), `skillsdir.go` `ErrNotFound`, and `runInit`'s `"skilldozer init:"` error-prefix strings. These ARE the developer/user docs for the exit-code and CLI contract.
  - P1.M2: the three `completions/*` files (they are the user-facing tab-completion surface).
- **Mode B (changeset-level, depends on the above):** `README.md` — update every command reference, **remove the now-obsolete "Reserved tag names" section** (there are no reserved names anymore), and document `--completions` as the eval/source install path. This is a distinct final task (P1.M2.T2).

---

## 4. Implementation plan

One phase, two milestones. M1 is the behavioral contract; M2 is the user-facing surfaces (completions + docs). M2 depends on M1 via the §14.4 lockstep invariant (completion files are frozen to `parseArgs`).

> **Reusable prior research:** `plan/003_3ace946c2a5c/architecture/external_deps.md` (embed verified, `//go:embed completions/_skilldozer` works without `all:`) and `code_change_map.md` (line-number conventions) still apply. The flag-conversion is NEW work not covered there.

### P1 — Subcommands → flags + skills-first completions

#### M1 — CLI contract conversion (`check`/`init`/`completion` → `--check`/`--init`/`--completions`)

The behavioral core. After M1, bare words are tags and the three actions are flags. All existing subsystems keep working; only the parsing/dispatch/error-message layer and tests change.

**T1 — `parseArgs` + `exclusivityError` + `usageText` + error strings (code)**
- **S1 (story points: 3):** In `main.go`:
  - **parseArgs:** delete the `case "check":`, `case "completion":`, and `case "init":` bare-token branches (and their reserved-subcommand positional-capture logic at `main.go:324-348`, including the `next == "init"` / `next != "check"` / `next != "completion"` special-casing — it is no longer needed). Add `case "--check":` → `c.check = true`; `case "--completions":` → `c.completion = true`; `case "--init":` → set `c.init = true` and, if the next token exists and does not start with `-`, capture it as `c.initStore` (mirroring the old `init <dir>` capture, now `--init <dir>`). Add `--init=<dir>` to the existing `=`-form switch (set `c.init=true`, `c.initStore=val`). The bare words `check`/`init`/`completions` now fall through to the `default:` branch and land in `c.tags` — that is the namespace-safety guarantee. `--store` still implies `init`; `--shell` still implies `completion` — leave that wiring.
  - **config struct doc comments** (`main.go:162-166`): update `check`/`init`/`completion`/`initStore` field comments to cite `--check` / `--init [<dir>]` / `--completions` (drop "subcommand" language).
  - **exclusivityError** (`main.go:782+`): the bool-driven *logic* survives (`c.check`/`c.init`/`c.completion` still gate the same families), but (a) rewrite every message so `'check'`→`'--check'`, `'init'`→`'--init'`, `'completion'`→`'--completions'`, and (b) add `c.completion` to the `init` block's mode set (currently `c.check || c.list || c.searchMode || c.all || c.path` omits it — it's caught by the completion block, but the message/init-set should be consistent) and add `--completions` to the init message's exclusion list.
  - **usageText** (`main.go:71-119`): rewrite USAGE to `skilldozer --check` / `skilldozer --init [<dir>]` / `skilldozer --completions [--shell <name>]`; update EXAMPLES (`skilldozer --check`, `skilldozer --init --store <dir>`, `eval "$(skilldozer --completions)"`); update OPTIONS rows to show `--check` / `--init [<dir>]` / `--completions [--shell <name>]` as long forms. Add the §6-header note that help/completions advertise long forms only.
  - **runInit error-prefix strings** (`main.go:1001-1110`, ~10 sites): `"skilldozer init: ..."` → `"skilldozer --init: ..."`.
  - **internal/skillsdir/skillsdir.go:275:** `ErrNotFound` message `"skilldozer is not configured; run \`skilldozer init\`"` → `"...run \`skilldozer --init\`"` (the §6.4/§13 acceptance greps for this).
  - Remove/rewrite any code comment that says a bare word is a "reserved" token or "subcommand" (~19 hits in non-test code). `runCompletion`/`completionScript`/`detectShell` **function names stay** (internal; renaming is cosmetic churn and adds risk) — only their doc comments move to `--completions`.
  - Verify: `go build ./... && go vet ./...`. The dispatch tests will fail until T2; that is expected (TDD).

**T2 — Update `main_test.go` to the flag contract (tests)**
- **S1 (story points: 3):** Flip every test that passes a bare `"check"`/`"init"`/`"completion"` token expecting `c.check`/`c.init`/`c.completion=true` to instead pass `"--check"`/`"--init"`/`"--completions"` (~66 functions / ~97 references — mechanical). Specifically: parseArgs tests, exclusivity tests (assert messages now mention `'--check'`/`'--init'`/`'--completions'`), and the `runInit`/`runCompletion` dispatch tests (invoke via `--init`/`--completions`). **Add new coverage for the namespace-safety guarantee:** a bare `check`/`init`/`completions` token now lands in `c.tags` (resolves as a tag or fails as unknown — never selects a mode), and `--init <dir>` / `--init=<dir>` capture the store dir. The `init <dir>` / `init --store` / duplicate-token edge cases (Issue 4 etc.) are re-expressed for the `--init` flag form. `go test ./...` must be fully green.

#### M2 — User-facing surfaces: completions rewrite + README sync

**T1 — Rewrite the three completion files (decision 20 + lockstep)**
- **S1 (story points: 2):** Edit `completions/skilldozer.bash`, `completions/_skilldozer`, `completions/skilldozer.fish` (these are the single source of truth and are `//go:embed`-ded, so a **rebuild** is required for `--completions` to emit the new bytes — §14.6):
  - **Remove** the `check`/`init`/`completion` first-positional subcommand offers in all three (the suppression walks, the `compadd ... check init completion`, the `__fish_is_first_arg -a 'check'/'init'/'completion'` directives). A bare `<tab>` now yields **skills only**.
  - **Flag matrix → long-form-only:** remove the short aliases (`-v`/`-h`/`-p`/`-l`/`-a`/`-f`/`-s`) from what is *offered* (bash `compgen -W` word list; zsh `_arguments` list; fish `-s` short opts). Keep them valid at runtime — they simply are not advertised.
  - **Add the new long flags:** `--check`, `--init`, `--completions` to each file's flag set. Route `--init <dir>` to directory completion (like `--store`), and `--completions --shell <name>` to the three-word enum `bash`/`zsh`/`fish`. `--store` routing is unchanged.
  - Tag completion (`skilldozer --relative --all 2>/dev/null`) stays byte-identical.
  - Update each file's `LOCKSTEP:` comment to cite decisions 19/20 and the long-form-only rule.
  - Acceptance: the §2 assertions (`EMBED_HAS_*`, `LONG_FORM_ONLY_*`) pass after `go build`.

**T2 — Sync README (Mode B changeset-level docs)**
- **S1 (story points: 1):** Depends on T1 above and M1. In `README.md`:
  - Update every command reference: `skilldozer init`→`skilldozer --init`, `skilldozer check`→`skilldozer --check`, `skilldozer completion`→`skilldozer --completions` (~15 sites: install, first-run, usage, adding-a-skill, how-it-finds-the-store, constraints).
  - **Remove the "Reserved tag names" paragraph** (~README.md:239-245) — it is now factually wrong: there are no reserved names, and that is the whole point of decision 19. Replace it (if useful) with a one-line note that bare words are always tags and actions are `--flags`.
  - Update the completions section: the eval/source one-liner is now `eval "$(skilldozer --completions)"` / `skilldozer --completions --shell fish | source`; describe skills-first + long-form-only behavior; note `--init`/`--check`/`--completions` as flags.
  - Add the §15 outline item 8 ("Shell completions") content if not present.
  - Verify: `grep -q 'skilldozer --completions' README.md && ! grep -q 'Reserved tag names' README.md`.

### Dependencies
- P1.M1.T2 (tests) → P1.M1.T1 (code).
- P1.M2.T1 (completions) → P1.M1.T1 (§14.4 lockstep: frozen to `parseArgs`).
- P1.M2.T2 (README) → P1.M1.T2 + P1.M2.T1 (describes both the new CLI and the new completions).

---

## 5. Guardrails (carry forward from §17, now with two new rows)

- ⚠️ **§14.4 lockstep:** when the flag set in `parseArgs` changes (this delta adds `--check`/`--init`/`--completions` and the completion files drop the short-form offers), update **all three** completion files identically. They are `//go:embed`-ded → **rebuild** before `--completions` reflects edits.
- ❌ **No bare-word subcommands** (decision 19, new §17 row): every non-tag action is a `--flag`. Do not re-introduce a reserved bare token.
- ⚠️ **Completions are skills-first and long-form-only** (decision 20, new §17 row): a bare `<tab>` shows skills; flag completion offers only `--long` forms, never short aliases.
- Keep `runCompletion` / `completionScript` / `detectShell` internal names as-is (only doc comments move to `--completions`) — renaming is unnecessary risk.
