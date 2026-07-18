# External Dependencies & PRD References

## Dependencies

The project has exactly ONE third-party dependency:
- `gopkg.in/yaml.v3 v3.0.1` — frontmatter parsing (PRD §7.3)

No new dependencies are needed for the `--link` batch feature. All required
functionality is in the Go standard library (`os`, `path/filepath`, `strings`,
`fmt`, `io`).

## PRD Section Cross-Reference for --link Batch

### §6.1 — CLI contract table
```
skilldozer --link <dir> [<dir>...] | Link one or more external skill directories
  into the store (see §8.4): pass --link once, then every following positional is
  a directory to link. | One absolute link path per successfully linked directory,
  in input order, one per line. | 0 if all link; 1 if any fail; 2 if no directory
  follows --link
```

### §6.3 — Default behavior / mutual exclusivity
- Mode flags are mutually exclusive with each other.
- `--link` accepts one or more positional directories.
- `--link` is the sole mode that **collects** trailing positionals.
- Mixing a `<tag>` with any non-`--link` mode flag is an error (exit 2).

### §6.4 — Error semantics for --link
- Each directory validated independently.
- Not a directory / store itself / inside it / no SKILL.md → stderr line naming dir.
- Name collision with non-symlink → stderr line naming dir.
- Directories processed in order; each success → stdout path, each failure → stderr line.
- Batch can yield mixed output (some stdout, some stderr).
- Exit 1 if any directory fails (successful links remain — idempotent).
- Single-directory case: one bad dir → nothing on stdout, exit 1.
- `--link` with no following directory → exit 2, nothing on stdout.
- Exact text: `skilldozer: --link requires at least one path to a skill directory`

### §8.4 — Linking behavior (full spec)
1. Resolve store via §8.3. Unconfigured → stderr hint, exit 1.
2. Collect link targets: `--link` consumes every following positional.
   `--link=<dir>` supplies the first; further positionals add to it.
   Zero → exit 2.
3. For each directory, in input order:
   a. Absolutize (expand ~, filepath.Abs).
   b. Validate: existing dir / not store-or-inside / HasSkillMD.
   c. Link name = filepath.Base(dir); link path = <store>/<name>.
   d. Conflict: nothing → create; existing symlink → refresh; non-symlink → refuse.
   e. Success → stdout link path + stderr "Linked … -> … (found via …)".
4. Exit 0 iff all succeed; exit 1 if any fail (successful links remain).

### §13 — Acceptance criteria for --link
```bash
# Single link
test "$(SKILLDOZER_SKILLS_DIR=/tmp/sd-link/store ./skilldozer --link /tmp/sd-link/src/linked)" = "/tmp/sd-link/store/linked"

# Multi-link: one --link, then several directories
out=$(SKILLDOZER_SKILLS_DIR=/tmp/sd-link/store ./skilldozer --link /tmp/sd-link/src/linked /tmp/sd-link/src/other)
printf '%s\n' "$out" | grep -qx '/tmp/sd-link/store/linked'
printf '%s\n' "$out" | grep -qx '/tmp/sd-link/store/other'

# Mixed batch: two valid + one invalid → valid link, exit 1, bad dir on stderr
out=$(... --link /tmp/sd-link/src/linked /tmp/sd-link/src/other /tmp/sd-link/notaskill 2>err); rc=$?
[ "$rc" = "1" ]
printf '%s\n' "$out" | grep -qx '/tmp/sd-link/store/linked'
grep -q 'notaskill' err

# Single bad dir: nothing on stdout, exit 1
out=$(... --link /tmp/sd-link/notaskill 2>/dev/null); rc=$?
[ -z "$out" ] && [ "$rc" = "1" ]

# No dir at all: exit 2
./skilldozer --link >/dev/null 2>&1; [ "$?" = "2" ]
```

### §14.1 — Completion behavior for multi-link
- `skilldozer --link <tab>` → Directories (first skill dir).
- `skilldozer --link d1 <tab>` → Directories (further skill dir; --link takes many).
- Rule 5: when `--link` is present, every following positional completes to dirs.

### Decisions 21 + 23
- Decision 21: --link creates symlinks, refreshes existing, refuses non-symlink.
- Decision 23: --link multi-target (batch). Non-atomic partial success. Mirrors
  `ln`/`git add` multi-target. Single-dir is degenerate n=1 batch.
