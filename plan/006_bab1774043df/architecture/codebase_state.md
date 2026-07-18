# Codebase State — `--link` Implementation Details

## Current `--link` Architecture (single-target)

### Config struct (`main.go:176-178`)
```go
link               bool     // `skilldozer --link <dir>` flag
linkTarget         string   // `--link <dir>` value (single directory)
linkMissingValue   bool     // --link / --link= seen with NO value → exit 2
```

### parseArgs — `=`-form (`main.go:266-274`)
```go
case "--link":
    c.link = true
    c.linkTarget = val
    if val == "" {
        c.linkMissingValue = true
    }
```

### parseArgs — next-token form (`main.go:403-429`)
```go
case "--link":
    if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
        c.link = true
        c.linkTarget = args[i+1]
        i++
    } else {
        c.linkMissingValue = true
    }
```
Key: a dashed follower (`--link --check`) is NOT consumed as the value →
`linkMissingValue = true`, `c.link` stays false. This is the "reject dashed
arguments as store/link values" feature (commit `ce01b55`).

### exclusivityError — `--link` block (`main.go:941-953`)
```go
if c.link {
    if hasTags {
        return true, "skilldozer: '--link' cannot be combined with tag arguments"
    }
    if c.check || c.init || c.completion || c.list || c.searchMode || c.all || c.path {
        return true, "skilldozer: '--link' cannot be combined with --check/--init/..."
    }
}
```
`hasTags` is computed from `len(c.tags) > 0` (`main.go:902`).

### runLink (`main.go:1424-1490`)
Single-target: resolve store → absolutize → validate (is dir / not store-inside
/ HasSkillMD) → conflict handling (create / refresh symlink / refuse non-symlink)
→ stdout=linkPath, stderr="Linked … -> … (found via …)". Returns 0 or 1.

### Missing-value guard (`main.go:608-610`)
```go
if c.linkMissingValue {
    fmt.Fprintln(stderr, "skilldozer: --link requires a path to a skill directory")
    return 2
}
```
Runs BEFORE exclusivityError.

### Dispatch (`main.go:651-655`)
```go
if c.link {
    return runLink(c, stdout, stderr)
}
```

## Existing Link Tests (must be updated for batch)

| Test | File:Line | Current expectation | Batch behavior |
|---|---|---|---|
| `TestParseArgsLinkNextToken` | main_test.go:3491 | `c.linkTarget == "/path/to/skill"` | `c.linkTargets[0]` |
| `TestParseArgsLinkEquals` | main_test.go:3508 | `c.linkTarget == "/path/to/skill"` | `c.linkTargets[0]` |
| `TestParseArgsLinkNoValue` | main_test.go:3523 | `c.link` false, `linkMissingValue` true | Same (zero dirs) |
| `TestParseArgsLinkEqualsEmpty` | main_test.go:3535 | `linkMissingValue` true | Same |
| `TestRunLinkNoValueExits2` | main_test.go:3549 | exit 2 | Same |
| `TestRunLinkEqualsEmptyExits2` | main_test.go:3564 | exit 2 | Same |
| `TestRunLinkWithTagsExits2` | main_test.go:3576 | exit 2 (tags+mode) | **CHANGES**: `sometag` is now a link target → validation failure → exit 1 |
| `TestRunLinkWithModeExits2` | main_test.go:3588 | exit 2 | Same |
| `TestRunLinkSuccess` | main_test.go:3602 | exit 0, one path | Same (n=1 batch) |
| `TestRunLinkRefreshSymlink` | main_test.go:3644 | exit 0 | Same |
| `TestRunLinkRefusesNonSymlink` | main_test.go:3671 | exit 1 | Same |
| `TestRunLinkRefusesNonSkill` | main_test.go:3704 | exit 1 | Same |
| `TestRunLinkRefusesFileTarget` | main_test.go:3728 | exit 1 | Same |
| `TestRunLinkRefusesInsideStore` | main_test.go:3747 | exit 1 | Same |
| `TestRunLinkUnconfiguredExits1` | main_test.go:3778 | exit 1 | Same |
| `TestRunLinkAbsolutizesRelativeTarget` | main_test.go:3797 | exit 0 | Same |
| `TestParseArgsLinkDashedFollowerNotConsumed` | main_test.go:3868 | exit 2 | Same |
| `TestParseArgsLinkDashedModifierFollowerIsMissingValue` | main_test.go:3889 | exit 2 | Same |
| `TestRunLinkDashedFollowerExits2NoMutation` | main_test.go:3946 | exit 2 | Same |

**Critical behavioral change**: `TestRunLinkWithTagsExits2` currently tests
`--link /tmp/foo sometag` → exit 2. With batch linking, `sometag` after
`--link` is collected as a link target (not a tag), so it becomes a validation
failure (not a directory) → exit 1. This test must be updated to reflect the
new semantics.

## Helper Functions Used by runLink

- `expandHome(s string) string` — expands `~`/`~/` to `$HOME` (defined in main.go)
- `skillsdir.Find() (dir, Source, err)` — resolves the store via §8.3 priority
- `skillsdir.HasSkillMD(dir string) bool` — checks for any `SKILL.md` at any depth
- `filepath.Abs`, `filepath.Base`, `filepath.Join`, `filepath.Clean` — path ops
- `os.Stat`, `os.Lstat`, `os.Symlink`, `os.Remove` — filesystem ops

## Completion Scripts — Current `--link` Handling

### bash (`completions/skilldozer.bash:38-44`)
```bash
case "$prev" in
    --search) return 0 ;;
    --store|--init|--link) COMPREPLY=($(compgen -d -- "$cur")); return 0 ;;
    --shell) COMPREPLY=($(compgen -W "bash zsh fish" -- "$cur")); return 0 ;;
esac
```
Only completes dirs when `$prev` is `--link`. After `--link d1`, `prev=d1`, falls
through to tag completion. **Needs fix**: detect `--link` anywhere in `words`.

### zsh (`completions/_skilldozer:49-51`)
```zsh
'--link[Link an external skill directory into the store]:directory:_files'
```
`_arguments` treats `--link` as single-value. Subsequent positionals get tags.
**Needs fix**: use `*:...` or a custom dispatcher for post-`--link` positionals.

### fish (`completions/skilldozer.fish:37-39`)
```fish
complete -c skilldozer -l link -d 'Link an external skill directory into the store' -r
```
`-r` makes the NEXT arg a file arg. After that, falls through to tag completion.
**Needs fix**: add a condition that offers dirs when `--link` has been seen.
