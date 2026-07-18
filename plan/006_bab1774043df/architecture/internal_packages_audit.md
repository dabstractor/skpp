# Internal Packages Audit — IMPLEMENTED vs PRD

Scope: every file under `/home/dustin/projects/skilldozer/internal/` mapped
against the authoritative PRD (`plan/006_bab1774043df/prd_snapshot.md`).
Each package section reports: what it does well, what's MISSING/PARTIAL, and
key function signatures. Severity legend: BLOCKER / GAP / NOTE.

Go module: `github.com/dabstractor/skilldozer` (`go 1.25`). Single third-party
dep confirmed: `gopkg.in/yaml.v3 v3.0.1` (matches PRD §7.3 "the only third-party
dependency"). Total: 4,455 lines across 18 files (8 source + 10 test).

---

## 1. `internal/skillsdir/skillsdir.go` (328 lines) — store locator

**PRD refs:** §8.1, §8.3 (resolution priority), §6.4 (vanished-store error).

### What it does well
- **All four §8.3 rules implemented** in exact PRD order, first-hit-wins, via
  private helpers `findEnv` / `findConfig` / `findSibling` / `findWalkUp`.
- **Config file IS supported** (§8.1, §8.3 rule 2). `findConfig()` delegates to
  `config.Path()` (which honors `SKILLDOZER_CONFIG`) and `config.Load()`; a
  relative `store` is resolved against the config file's own dir (not cwd).
- **`SKILLDOZER_CONFIG` IS supported** — transitively, through the
  `internal/config` package's `config.Path()` (see §6 below). `skillsdir`
  itself reads `SKILLDOZER_SKILLS_DIR` directly (rule 1).
- **"Which rule won" reporting (§8.3):** `Source` enum + `Source.String()`
  emits exactly the four PRD labels: `SKILLDOZER_SKILLS_DIR`, `config file`,
  `sibling of binary`, `ancestor of cwd`. `Find()` returns the `Source` so
  `main --path` can print it.
- **Vanished-store error (§6.4):** if config names a non-existent `store:`,
  `findConfig` returns a `vanishedStore` signal and `Find` wraps it as
  `ErrConfiguredStoreMissing` instead of falling through — preventing silent
  load of an unrelated sibling/walk-up store. This is a strict superset of the
  PRD's "fall through" wording and a *correct* refinement.
- **Sibling rule (§8.2):** symlink-aware via `os.Executable()` +
  `filepath.EvalSymlinks()` (in `resolveSiblingFromExe`), factored out so the
  logic is unit-testable independent of the test binary's own path.
- **Walk-up rule (§8.3):** checks `skills/` subdir that contains at least one
  `SKILL.md` (via exported `HasSkillMD`), matching PRD §8.3's "at least one
  SKILL.md" qualification. Ascent stops cleanly at the filesystem root.
- **`HasSkillMD` exported** — doubles as the §8.2 cwd-auto-detect predicate
  for `--init`.

### What's MISSING / PARTIAL
- NOTE: `--init`'s "report which rule won" (§8.2 step 5) lives in `main.go`,
  not here — `skillsdir` only *provides* the `Source`; it does not print. Out
  of scope for this package, called out so the parent does not look for it here.
- NOTE: env rule 1 does NOT `EvalSymlinks` the env path (documented as
  intentional — the user points exactly where they want). The walk/cycle guard
  in `discover` handles canonicalization downstream, so this is consistent.

### Key signatures
```go
func Find() (dir string, src Source, err error)         // PRD §8.3 entry point
func HasSkillMD(dir string) bool                         // §8.2 cwd-auto-detect predicate
type Source int                                          // enum: SourceEnv|SourceConfig|SourceSibling|SourceWalkUp
func (s Source) String() string                          // §8.3 labels for --path
var ErrNotFound = errors.New("skilldozer is not configured; run `skilldozer --init`")
var ErrConfiguredStoreMissing = errors.New("configured skills store directory does not exist")
// (private) func findEnv() / findConfig() / findSibling() / findWalkUp()
// (private) func resolveSiblingFromExe(exe string) (string, bool)
// (private) func findWalkUpAncestor(start string) (string, bool)
```

### Verdict
**Complete.** All four precedence rules, the `--path` labels, the vanished-store
refinement, and the cwd-auto-detect predicate are present and tested (700-line
test file covers precedence matrix, symlinks, vanished stores, env-vs-config
conflicts). No gaps against PRD.

---

## 2. `internal/discover/` (discover.go + index.go + skill.go, 398 lines) — catalog walk

**PRD refs:** §7.1 (discovery), §7.3 (frontmatter parsing), §2 constraint 1
(manifest-free).

### `discover.go` (119 lines) — frontmatter model + parser
- **Uses `gopkg.in/yaml.v3`** (the one allowed dep) — confirmed in source.
- **`Frontmatter` struct** models the Agent Skills spec fields: `Name`,
  `Description`, `License`, `Compatibility`, `Metadata map[string]any`,
  `AllowedTools` (space-delimited string per spec), `DisableModelInvocation`,
  plus `HasFM bool` (yaml:"-", records whether a `---` block existed).
- **`ParseFrontmatter(path) (Frontmatter, body, err)`** implements §7.3:
  - Strips a leading UTF-8 BOM.
  - Handles CRLF `\r` in fence-line comparison.
  - No `---` block ⇒ `{HasFM:false}`, whole file as body, nil error (lenient —
    skill still resolves by directory; `--check` flags it).
  - Opening fence but no closing fence ⇒ treated as no frontmatter (lenient).
  - Valid fences with **broken YAML** ⇒ HARD error returned (lenient about
    unknown *keys* only, not corrupt YAML — `--check` relies on this).
  - Returns body VERBATIM (does not trim folded-scalar trailing newline; the
    `check` length test trims if it wants visible length).

### `index.go` (161 lines) — recursive walk
- **Walks recursively** via a hand-rolled `os.ReadDir` recursion (NOT
  `filepath.WalkDir`). Nested skills count: `writing/reddit/SKILL.md` is a skill
  with `RelTag="writing/reddit"`.
- **Follows symlinks** — explicitly. Each entry is run through
  `filepath.EvalSymlinks`; linked dirs are recursed into and reported under
  their on-disk *link name* (`displayPath`), not the resolved target, matching
  §7.1 "reported under its on-disk link name".
- **Breaks cycles** via a `visited` set of canonical real paths
  (`EvalSymlinks` output) — a link back at an ancestor resolves to a string
  already in the set and is skipped. Bind-mount inode cycles are explicitly
  out of scope (documented).
- **relTag normalization (§7.2 step 1):** `filepath.Rel(root, displayPath)`
  then `filepath.ToSlash` — `/`-separated on every platform.
- **Manifest-free (§2):** no index file; rebuilt from disk every call.
- **Sorts by `RelTag`** for deterministic `--all` output (§6.1 "sorted by tag").
- **Error policy:** missing/unreadable/non-dir root ⇒ returned error; per-entry
  errors skipped; malformed-YAML SKILL.md ⇒ lenient `HasFM=false` Skill built so
  the skill is still resolvable by directory (the parse error is dropped here;
  `check` re-parses `s.SourceFile` to recover it).
- **Two-coordinate threading** (`displayPath` for tags/Dir, `realPath` for reads
  + cycle key) is the non-obvious correctness hinge and is well documented.

### `skill.go` (118 lines) — Skill type + metadata extraction
- **`Skill` struct** carries §7.1 capture fields: `Dir`, `RelTag`, `Name`,
  `Description`, `Keywords []string`, `Category`, `Aliases []string`, `HasFM`,
  `SourceFile` (== `Dir/SKILL.md`). No yaml tags (built, not unmarshaled).
- **`BuildSkill(dir, relTag, Frontmatter) Skill`** — the S1→T5 boundary; total
  (never errors/panics, even on nil metadata map).
- **`toStringSlice(any) []string`** normalizes yaml.v3's `[]any` into `[]string`
  for Keywords/Aliases: nil→nil; `[]any`→strings (non-strings silently dropped);
  single string→`[s]` (lenient: `keywords: writing` → `["writing"]`).

### What's MISSING / PARTIAL
- NOTE: A directory literally named `SKILL.md` is skipped (handled), and a
  symlink-to-dir named `SKILL.md` is also skipped — matches §7.1 "directory
  that directly contains a SKILL.md" (a SKILL.md *dir* is not a SKILL.md file).
- NOTE: Broken/unresolvable symlinks are skipped silently (Stat/EvalSymlinks
  error → `continue`). This is lenient and reasonable; PRD does not require
  surfacing them.
- GAP (minor, by design): the walk does not detect "grouping dir with no
  SKILL.md" — only dirs *containing* SKILL.md are indexed. This is intentional
  (a heuristic would false-positive on legitimate grouping dirs) and is what
  forces the §9 reframing in `check` (see §5). Documented, not a bug.

### Key signatures
```go
func ParseFrontmatter(path string) (fm Frontmatter, body string, err error)  // §7.3
func Index(skillsDir string) ([]Skill, error)                                // §7.1 walk
func BuildSkill(dir, relTag string, fm Frontmatter) Skill                    // metadata extraction
type Frontmatter struct{ ...; Metadata map[string]any `yaml:"metadata,omitempty"`; HasFM bool `yaml:"-"` }
type Skill struct{ Dir, RelTag, Name, Description string; Keywords, Aliases []string; Category string; HasFM bool; SourceFile string }
```

### Verdict
**Complete and faithful.** Recursive walk, symlink-following, cycle-breaking,
yaml.v3 parsing, BOM/CRLF handling, relTag normalization, and the
manifest-free contract are all implemented. 905 lines of tests across the three
test files cover nested skills, symlinks, cycles, BOM/CRLF, broken YAML, and
edge cases.

---

## 3. `internal/resolve/resolve.go` (195 lines) — tag → skill precedence

**PRD refs:** §7.2 (precedence), §6.4 (error semantics).

### What it does well
- **All four §7.2 precedence steps implemented** in exact order, first-match-
  wins, with later steps consulted only if every earlier step produced nothing:
  1. **Canonical** — `tag == skill.RelTag` (case-sensitive). Exact-and-unique;
     no ambiguity possible (inlined loop).
  2. **Basename** — `tag == final '/'-component of RelTag`. `>1` ⇒ ambiguous.
  3. **Name** — `tag == skill.Name`; **guards on `Name != ""`** so a
     missing-name skill and an empty tag cannot spuriously match.
  4. **Alias** — `tag ∈ skill.Aliases`; empty alias never matches.
  5. Otherwise ⇒ `*UnknownError`.
- **Ambiguity short-circuits (§6.4):** `*AmbiguousError` returned immediately
  at the first ambiguous step; later, looser steps are NOT consulted.
- **`AmbiguousError.Candidates` is sorted** for deterministic stderr output
  regardless of input slice order (§6.4 "stable stderr for scripting").
- **Pure function:** takes `[]discover.Skill` as a param — no filesystem, no
  globals, no I/O. `main` supplies the index.
- **`MatchKind` enum + `String()`** exported (Canonical/Basename/Name/Alias) —
  lets callers annotate which step resolved a tag (e.g. debug/`--list`).

### What's MISSING / PARTIAL
- None against PRD §7.2/§6.4. The field-level guards (Name≠"", alias≠empty) are
  a *correct* hardening beyond the literal PRD text.

### Key signatures
```go
func Resolve(tag string, skills []discover.Skill) (Result, error)            // §7.2 entry
type Result struct{ Skill discover.Skill; Match MatchKind }
type MatchKind int                                                          // Canonical|Basename|Name|Alias
type UnknownError struct{ Tag string }                                      // §6.4 unknown tag
type AmbiguousError struct{ Tag string; Candidates []string }               // §6.4 ambiguous
// (private) func collectMatches(skills, pred) []discover.Skill
// (private) func basename(relTag string) string
// (private) func sortedRelTags(skills) []string
```

### Verdict
**Complete.** All four steps, correct ambiguity handling, deterministic
candidates, pure-function design. 288-line test file.

---

## 4. `internal/search/search.go` (92 lines) — `--search` filter

**PRD refs:** §6.1 (`--search`), §10 (metadata conventions). **Tension between
the two is the key finding.**

### What it does well
- **Case-insensitive substring** over a clearly enumerated field set.
- **Query lowercased once** (per-call), fields lowercased lazily per `Contains`.
- **Keywords and Aliases matched INDIVIDUALLY** (not `strings.Join`'d) — a
  query spanning a boundary between two keywords/aliases must not match. This
  is a deliberate correctness detail (avoids false positives like "wriocial").
- **Empty query matches everything** — `--search ""` behaves like `--list`
  (exit 1 only if store empty). Documented as natural substring semantics.
- **Pure function** over `[]discover.Skill`; preserves input order (caller's
  `discover.Index` sort).
- **Frontmatter-less skill still discoverable** by its `RelTag` (Name/Desc
  empty, Keywords nil, but RelTag always present) — consistent with `resolve`.

### ⚠️ FIELD SET — divergence between PRD §6.1 and §10 (the headline finding)
- **PRD §6.1 table** (`--search <q>`): "Substring (case-insensitive) search
  over **tag, frontmatter `name`, `description`, and `metadata.keywords`**"
  → **4 fields** (tag, name, description, keywords). **Aliases and category
  are NOT in the §6.1 list.**
- **PRD §10** states keywords/category/aliases "exist only to enrich
  `skilldozer --search`" → implies **6 fields**.
- **Implementation searches 6 fields:** `RelTag`, `Name`, `Description`, each
  `Keyword`, each `Alias`, AND `Category`.
- The source documents this as a **deliberate decision** ("decisions.md §D4:
  §10 wins over §6.1's summary field list"). The wider field set also keeps
  `--search` consistent with `resolve` step 4 (alias resolution).

**Severity: NOTE (deliberate, documented, defensible).** If a reviewer reads
§6.1 alone they will expect 4 fields and see 6; the §10 justification resolves
the conflict, but this is worth flagging as the single most likely "is this a
bug?" question. No code change needed unless the product owner wants strict
§6.1 compliance (drop aliases + category from search).

### What's MISSING / PARTIAL
- None functional. The field-set choice is the only thing to call out (above).

### Key signatures
```go
func Search(query string, skills []discover.Skill) []discover.Skill   // §6.1/§10 entry
// (private) func matches(q string, s discover.Skill) bool
```

### Verdict
**Complete.** 6-field substring search (vs PRD §6.1's 4-field summary —
intentional §10 enrichment, documented in-source). 170-line test file.

---

## 5. `internal/check/check.go` (250 lines) — `--check` validation

**PRD refs:** §9 (validation rules), §3 (Agent Skills name rules).

### What it does well
- **Implements nearly all §9 rules:**
  - ERROR: frontmatter missing `name` or `description`, or `description` empty.
  - ERROR: `name` violates Agent Skills rules — **charset/structure via regex**
    `^[a-z0-9]+(-[a-z0-9]+)*$` (lowercase a-z0-9, single hyphens, no
    leading/trailing/consecutive hyphens) AND **64-char max** (`nameLenMax=64`,
    checked separately because regex can't express length).
  - ERROR: duplicate frontmatter `name` across skills (global pass, "also in"
    list sorted for determinism; one ERROR per owner).
  - WARN: `description` > 1024 chars (`descLenMax=1024`).
- **Malformed-YAML recovery (the non-obvious part):** because `discover.Index`
  drops the parse error, `check` RE-RUNS `discover.ParseFrontmatter(s.SourceFile)`
  to recover the malformed-YAML-vs-no-block distinction. Three cases handled:
  - `perr != nil` (broken YAML / file vanished) → ERROR "invalid SKILL.md
    frontmatter: ..." (this is the **reframed** §9 "skill dir has no SKILL.md",
    since discover only ever emits dirs *containing* SKILL.md).
  - `!fm.HasFM` (no `---` block) → ERROR "missing frontmatter block".
  - else → field checks via `checkFields`.
- **Structured `Report`** (`BySkill []SkillReport`, `Errors`, `Warnings` counts)
  with `HasErrors()` for the exit-code decision (§9: exit 1 iff any ERROR;
  WARNs never affect exit code).
- **Description length measured on TRIMMED value** (`strings.TrimSpace`),
  matching `ui`'s display length — a whitespace-only description trims to `""`
  → ERROR (not WARN).
- **Severity enum + String()** (`OK`/`WARN`/`ERROR`).

### What's MISSING / PARTIAL
- **§9 WARN "skill dir is empty besides `SKILL.md`" is INTENTIONALLY NOT
  implemented.** The source explicitly justifies this: the shipped example skill
  IS only `SKILL.md`, and enabling the WARN would break the §13 acceptance
  ("reports the example as OK"). **Severity: NOTE (deliberate, PRD §9 itself
  marks it "optional").** No code change needed.
- **§9 ERROR "skill dir has no `SKILL.md`" is REFRAMED** to "invalid SKILL.md
  frontmatter" / "missing frontmatter block" because `discover.Index` only
  emits dirs that contain a `SKILL.md`, so the literal rule can never fire. A
  grouping dir without SKILL.md is never indexed/inspected. The reframing is
  documented in-source. **Severity: NOTE (the actionable form of the rule).**
  If a reviewer insists on the literal §9 wording, this is a coverage gap by
  construction, but the documented rationale (false-positive risk on grouping
  dirs) is sound.
- The regex cannot express the 64-char max — handled separately (`len > 64`
  check before regex). Not a gap, just a noted two-part check.

### Key signatures
```go
func Check(skills []discover.Skill) Report                                  // §9 entry
func (r Report) HasErrors() bool                                            // exit-code decision
type Severity int                                                          // LevelOK|LevelWarn|LevelError
type Finding struct{ Level Severity; Message string }
type SkillReport struct{ Skill discover.Skill; Findings []Finding }
type Report struct{ BySkill []SkillReport; Errors, Warnings int }
// (private) var validName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
// (private) func checkFields(fm Frontmatter) []Finding
// (private) func appendDupFindings(rep *Report, owners []nameOwner)
// (private) const nameLenMax = 64; const descLenMax = 1024
```

### Verdict
**Complete (with two deliberate, documented omissions).** All ERROR rules and
the description WARN are implemented; the optional "empty besides SKILL.md"
WARN is intentionally skipped (would break §13 acceptance); the "no SKILL.md"
ERROR is reframed to the actionable "invalid/missing frontmatter" form. 234-line
test file.

---

## 6. `internal/config/config.go` (159 lines) — settings sidecar read/write

**PRD refs:** §8.1 (config file), §8.2 (default store), §2 constraint 1
(settings ≠ catalog).

### What it does well
- **Reads/writes the YAML settings file** — `Load(path)` and `Save(path, File)`.
  `File` has a single `Store string` field (the store location) with
  `yaml:"store,omitempty"`.
- **Settings sidecar, NOT a catalog (§2/§17):** enforced by construction — the
  struct only carries `store`. The doc comment repeatedly warns against growing
  it into a catalog.
- **Lenient on unknown keys** (yaml.v3 default, no `KnownFields(true)`),
  forward-compatible. **Hard error on broken YAML** — returned as-is.
- **`Path()` resolves the config-file location (§8.1):**
  - `SKILLDOZER_CONFIG` non-empty ⇒ taken as the LITERAL path
    (`filepath.Clean`'d, NOT joined to config home — relative values work for
    tests/multiple profiles, per §8.1).
  - Otherwise `$XDG_CONFIG_HOME/skilldozer/config.yaml` via
    `os.UserConfigDir()` (honors `XDG_CONFIG_HOME`, falls back to `~/.config`,
    rejects a relative `XDG_CONFIG_HOME` with an error).
- **`DefaultStore()` resolves the out-of-the-box store (§8.2/§8.3):**
  - Absolute `XDG_DATA_HOME` ⇒ `$XDG_DATA_HOME/skilldozer/skills`.
  - Otherwise `~/.local/share/skilldozer/skills` via `os.UserHomeDir()`.
  - A *relative* `XDG_DATA_HOME` is INVALID (guarded by `filepath.IsAbs`) and
    ignored — never produces a relative store path.
- **`Save` is deterministic:** non-empty Store writes exactly `store: <value>\n`
  (struct-field order, not sorted); creates parent dir with `MkdirAll(0o755)`,
  writes file `0o644`.
- **`Load` returns the read error VERBATIM** (not wrapped) so callers can
  `errors.Is(err, fs.ErrNotExist)` to distinguish missing-vs-broken —
  `findConfig` relies on this to fall through.
- **`Path`/`DefaultStore` are pure functions of the environment** (no FS reads)
  — locating vs. using is cleanly separated.

### What's MISSING / PARTIAL
- None against PRD §8.1/§8.2. The "unknown keys ignored" forward-compatibility
  matches §8.1 ("room to grow: default category, color prefs, etc.").

### Key signatures
```go
func Load(path string) (File, error)              // §8.1 read (verbatim read error)
func Save(path string, f File) error              // §8.1 write (deterministic, MkdirAll)
func Path() (string, error)                       // §8.1 file location (SKILLDOZER_CONFIG / XDG)
func DefaultStore() (string, error)               // §8.2/§8.3 default store (XDG_DATA_HOME / ~/.local/share)
type File struct{ Store string `yaml:"store,omitempty"` }
// (private) const configEnv = "SKILLDOZER_CONFIG"
```

### Verdict
**Complete.** Read/write of the settings sidecar, config-file path resolution
(with `SKILLDOZER_CONFIG`), default-store resolution, and the
settings-≠-catalog discipline are all present and tested (283-line test file
covers absolute/relative/empty `SKILLDOZER_CONFIG`, relative `XDG_CONFIG_HOME`
rejection, `XDG_DATA_HOME` honoring).

---

## 7. `internal/ui/ui.go` (178 lines) — `--list`/`--search` table

**PRD refs:** §6.1 (`--list` table), §6.2 (`--no-color`).

### What it does well
- **TAG / NAME / DESCRIPTION table** exactly as §6.1 specifies. Column rules:
  - TAG = `Skill.RelTag` (the canonical `/`-normalized tag).
  - NAME = `Skill.Name`, or `"(none)"` when empty.
  - DESCRIPTION = `Skill.Description`, or `"(missing)"` when `HasFM==false` OR
    the description is empty/blank (matches §7.3 "(missing)" convention).
- **Word-wrap** to `descWrapWidth=40`; continuation lines leave TAG/NAME cells
  blank (spaces) so columns stay aligned.
- **ANSI color is opt-in** via a `useColor bool` parameter (caller owns the
  TTY/`--no-color` decision). Color: **bold header**, **cyan TAG**;
  NAME/DESCRIPTION default. `ansiReset` appended after every colored run so no
  bleed. Deterministic for tests (no real terminal needed).
- **Dynamic column widths** computed from PLAIN content (independent of color),
  and padding is applied to PLAIN text BEFORE `paint` so ANSI escape bytes do
  not corrupt the `len()`-based padding math.
- **`displayWidth` uses rune count** (`utf8.RuneCountInString`), so multi-byte
  runes (`é`, `—`, smart quotes) pad correctly for the common case.
- **Empty slice prints nothing** — `main` exits 1 "if no skills found" before
  calling this (PrintList is defensive, not authoritative).
- **`--search` reuses `PrintList`** (PRD §6.1 "same table format as --list")
  — the caller passes a filtered, still-sorted slice; `PrintList` does not
  re-sort.

### What's MISSING / PARTIAL
- NOTE: **No terminal-width detection** (`TIOCGWINSZ`/`x/term`) — deliberately
  avoided to keep `yaml.v3` the only third-party dep (PRD §4/§7.3). Wrap width
  is fixed at 40. Documented as a deliberate trade-off for deterministic,
  testable, dependency-free output. **Severity: NOTE (deliberate).**
- NOTE: **CJK wide runes count as 1 column**, not 2 — a full East-Asian width
  table would be needed for exact alignment. Deliberately omitted
  (dependency-free). Documented. **Severity: NOTE (deliberate, minor).**
- NOTE: `PrintList` is the only public function — there is no separate
  `PrintSearch`; `--search` reuses `PrintList` directly (correct per PRD).

### Key signatures
```go
func PrintList(w io.Writer, skills []discover.Skill, useColor bool)   // §6.1 table
// (private) const descWrapWidth = 40
// (private) const ansiReset / ansiBold / ansiCyan
// (private) func displayWidth(s string) int          // rune count
// (private) func padRight(s string, n int) string
// (private) func wrapWords(s string, width int) []string
```

### Verdict
**Complete (with two deliberate, documented trade-offs).** The §6.1 table,
`(missing)`/`(none)` conventions, word-wrap, opt-in color, and color-safe
padding are all implemented. Fixed wrap width and rune-count width are
deliberate dependency-free choices. 275-line test file.

---

## Cross-cutting summary

| Package | PRD coverage | Gaps | Severity |
|---|---|---|---|
| `skillsdir` | §8.1/§8.3/§6.4 — all 4 rules + labels + vanished-store | none | — |
| `discover` | §7.1/§7.3/§2 — recursive, symlink-following, cycle-broken, yaml.v3 | none (grouping-dir non-detection is by design) | — |
| `resolve` | §7.2/§6.4 — all 4 steps, ambiguity short-circuit | none | — |
| `search` | §6.1/§10 — 6-field substring (§10 enrichment) | §6.1 summary lists only 4 fields | **NOTE** (deliberate, documented) |
| `check` | §9/§3 — all ERRORs + description WARN | optional "empty besides SKILL.md" WARN skipped; "no SKILL.md" ERROR reframed | **NOTE** (both deliberate, documented) |
| `config` | §8.1/§8.2 — read/write settings sidecar, SKILLDOZER_CONFIG, default store | none | — |
| `ui` | §6.1/§6.2 — TAG/NAME/DESC table, word-wrap, opt-in color | no term-width detection; CJK width≈rune count | **NOTE** (both deliberate, documented) |

### Residual risks
- **The only finding worth a product decision** is `search`'s 6-field vs §6.1's
  4-field list. It is documented as intentional (§10 wins over §6.1) and keeps
  `--search` consistent with `resolve`'s alias step. If strict §6.1 compliance
  is wanted, drop Aliases + Category from `matches()`. No code change needed
  otherwise.
- All other gaps are deliberate, documented trade-offs (dependency-free UI,
  optional WARN skipped to satisfy §13, reframed §9 ERROR). No BLOCKERs found.
- The `skillsdir` vanished-store refinement (`ErrConfiguredStoreMissing`) is a
  *stricter* behavior than §8.1's literal "fall through" — verify the parent is
  aware `main` must surface this error rather than treat it as "unconfigured".

### Where to start
Open `internal/skillsdir/skillsdir.go` first — it is the §8 resolution engine
(the PRD's "hardest part; do first" per §18) and the gateway the rest of the
tool depends on (`discover`/`resolve`/`check` all need a store dir to walk).
`Find()` at the bottom is the public entry; the per-rule helpers above it are
the implementation.
