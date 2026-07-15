# Code Change Map — Delta 004 (main.go)

Exact line numbers verified against HEAD `f30d5c5` (1288 lines). All changes are
within `main.go` and `internal/skillsdir/skillsdir.go`.

---

## Change Group 1: parseArgs flag conversion (main.go:179-369)

### 1a. Delete bare subcommand cases

**`case "check":` — main.go:284-293**
```go
// DELETE THIS ENTIRE CASE:
case "check":
    c.check = true
```
Replace with flag form (see 1b).

**`case "completion":` — main.go:294-299**
```go
// DELETE THIS ENTIRE CASE:
case "completion":
    c.completion = true
```
Replace with `--completions` flag form.

**`case "init":` — main.go:324-351**
```go
// DELETE THIS ENTIRE CASE (including the Issue-4 duplicate-init logic,
// next != "check" / next != "completion" special-casing):
case "init":
    c.init = true
    if i+1 < len(args) { ... }
```
Replace with `--init` flag form.

### 1b. Add flag-form cases

In the main `switch a {` (same block, lines ~252+), ADD:

```go
case "--check":
    c.check = true
case "--completions":
    c.completion = true
case "--init":
    c.init = true
    if i+1 < len(args) {
        next := args[i+1]
        if !strings.HasPrefix(next, "-") {
            c.initStore = next
            i++
        }
        // else: a dashed flag (--init --store …) → left for its own case
    }
```

### 1c. Add `--init=<dir>` to the `=`-form switch (main.go:196-234)

After `case "--store":` block (line 218-227), ADD:
```go
case "--init":
    c.init = true
    c.initStore = val
    if val == "" { c.storeMissingValue = true }
```
NOTE: mirrors the existing `--store=` form. `--init=` with empty value triggers
`storeMissingValue` (same exit-2 guard as `--store=`).

### 1d. Bare words now fall through to `default:` (main.go:353-368)

The `default:` branch already captures non-dashed tokens into `c.tags`. After
deleting the bare-subcommand cases, `check`/`init`/`completions` tokens land
here automatically — that IS the namespace-safety guarantee. No code change
needed in `default:` itself.

### 1e. `--store` still implies init; `--shell` still implies completion

These are ALREADY wired (main.go:218-233 `=`-form, main.go:300-323 token form).
Do NOT change this wiring. `--store <dir>` → `c.init=true` + `c.initStore=val`;
`--shell <name>` → `c.completion=true` + `c.completionShell=val`.

---

## Change Group 2: config struct doc comments (main.go:148-168)

Update the doc comments on these fields (LOGIC unchanged, only comments):

| Line | Field | Old comment | New comment |
|------|-------|-------------|-------------|
| 162 | `check` | `\`skilldozer check\` subcommand: validate every skill...` | `\`skilldozer --check\` flag: validate every skill...` |
| 163 | `init` | `\`skilldozer init [<dir>]\` first-run setup...` | `\`skilldozer --init [<dir>]\` first-run setup...` |
| 164 | `initStore` | `init <dir> positional or --store <dir>...` | `--init <dir> flag or --store <dir>...` |
| 166 | `completion` | `\`skilldozer completion\` subcommand (§14.6)...` | `\`skilldozer --completions\` flag (§14.6)...` |

---

## Change Group 3: exclusivityError messages (main.go:782-835)

Rewrite ALL error messages. The bool-driven **logic** is unchanged (same gating);
only message strings and the init-block mode set change.

### Messages to rewrite (old → new):

| Line | Old message | New message |
|------|-------------|-------------|
| 804 | `'check' cannot be combined with tag arguments` | `'--check' cannot be combined with tag arguments` |
| 807 | `'check' cannot be combined with --path/--list/--search/--all` | `'--check' cannot be combined with --path/--list/--search/--all` |
| 815 | `'init' cannot be combined with tag arguments` | `'--init' cannot be combined with tag arguments` |
| 818 | `'init' cannot be combined with --list/--search/--all/--path/check` | `'--init' cannot be combined with --check/--list/--search/--all/--path` |
| 825 | `'completion' cannot be combined with tag arguments` | `'--completions' cannot be combined with tag arguments` |
| 829 | `'completion' cannot be combined with check/init/--path/--list/--search/--all` | `'--completions' cannot be combined with --check/--init/--path/--list/--search/--all` |

### Mode-set consistency fix (line 818):

The init-block mode set at line 818 (`{c.check, c.list, c.searchMode, c.all, c.path}`)
currently OMITS `c.completion`. The completion-block at 829 backstops it, but the
delta plan requests consistency. ADD `c.completion` to the init-block's mode set:
```go
if c.check || c.completion || c.list || c.searchMode || c.all || c.path {
    return true, "skilldozer: '--init' cannot be combined with --check/--completions/--list/--search/--all/--path"
}
```
Similarly, add `c.completion` to the check-block mode set at line 807 for consistency.

---

## Change Group 4: usageText constant (main.go:71-117)

### USAGE block (lines 80-82):
```
# OLD:                              # NEW:
skilldozer check                    skilldozer --check
skilldozer init [<dir>]             skilldozer --init [<dir>]
skilldozer completion [--shell <n>] skilldozer --completions [--shell <name>]
```

### EXAMPLES block (lines 95-97):
```
# OLD:                              # NEW:
skilldozer check                    skilldozer --check
skilldozer init --store <dir>       skilldozer --init --store <dir>
eval "$(skilldozer completion)"     eval "$(skilldozer --completions)"
```

### OPTIONS block (lines 104-108):
```
# OLD:                              # NEW:
check          Validate...          --check          Validate every skill on disk...
init [<dir>]   First-run setup...   --init [<dir>]   First-run setup: pick/create the skills store...
completion     Emit...              --completions [--shell <name>]  Emit the shell completion script for eval (§14.6)
```

### Add §6-header note about long-form-only advertising:
Add a line after USAGE (or before OPTIONS) noting that `--help` and `--completions`
advertise long forms only; short aliases (-a, -l, -s, -f, -p, -h, -v) remain valid
for typing but are not advertised.

---

## Change Group 5: Error prefix strings

### 5a. skillsdir.ErrNotFound — `internal/skillsdir/skillsdir.go:275`

```go
// OLD:
var ErrNotFound = errors.New("skilldozer is not configured; run `skilldozer init`")
// NEW:
var ErrNotFound = errors.New("skilldozer is not configured; run `skilldozer --init`")
```
This is the ONLY change to `internal/skillsdir/skillsdir.go`.

### 5b. runInit/setupStore error prefixes — main.go (8 sites)

ALL occurrences of `"skilldozer init: "` → `"skilldozer --init: "`:

| Line | String |
|------|--------|
| 1001 | `"skilldozer --init: resolve cwd: %w"` |
| 1005 | `"skilldozer --init: resolve default store: %w"` |
| 1027 | `"skilldozer --init: absolutize store: %w"` |
| 1090 | `"skilldozer --init: create store dir %q: %w"` |
| 1095 | `"skilldozer --init: read store dir %q: %w"` |
| 1100 | `"skilldozer --init: create example dir: %w"` |
| 1103 | `"skilldozer --init: seed example SKILL.md: %w"` |
| 1110 | `"skilldozer --init: write config %q: %w"` |

### 5c. Comment-only references (3 sites)

| Line | Context |
|------|---------|
| 1086 | setupStore doc: `"skilldozer init: <step>: %w"` → `"skilldozer --init: <step>: %w"` |
| 1150 | runInit comment: `resolveStore wraps with "skilldozer init: …"` → `"--init"` |
| 1162 | runInit comment: `setupStore wraps with "skilldozer init: …"` → `"--init"` |

---

## Change Group 6: Function doc comments (no logic change)

### 6a. completionScript — main.go:1112-1120
- Doc comment references `skilldozer completion` → update to `skilldozer --completions`

### 6b. detectShell — main.go:1234-1243
- Doc comment references `skilldozer completion` → update to `skilldozer --completions`

### 6c. runCompletion — main.go:1257-1271
- Doc comment references `skilldozer completion` (multiple sites) → update to `skilldozer --completions`
- The `eval "$(skilldozer completion)"` example in the doc → `eval "$(skilldozer --completions)"`

### 6d. Other code comments (~16 sites)
Remove/rewrite comments containing "reserved" or "subcommand":
- Lines 10, 162, 166, 285, 288, 295, 325, 328, 330, 631, 821, 1191
- Replace "subcommand" with "flag"; replace "reserved positional token" with "flag"

### 6e. Package doc (main.go:10)
- "subcommands like `check`" → "flags like `--check`"

---

## Dispatch verification

The `run()` function dispatches to modes based on config fields:
- `c.check` → `runCheck()`
- `c.init` → `runInit()`
- `c.completion` → `runCompletion()`

These dispatch blocks are UNCHANGED — the config fields still drive the same
functions. Only how those fields get SET changes (bare tokens → flags).
