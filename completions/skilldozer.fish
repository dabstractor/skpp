# Fish completion for skilldozer.
#
# Install:
#   cp completions/skilldozer.fish ~/.config/fish/completions/skilldozer.fish
#
# Tags are derived DYNAMICALLY from disk by calling `skilldozer --relative --all`
# (skilldozer is manifest-free, PRD §2.1: there is no sidecar catalog to read).
#
# LOCKSTEP: the flag list below is frozen to `main.go parseArgs()`. If a future
# task adds/renames a flag there, update this file — and the bash/zsh files —
# identically.

# No file completion: skilldozer takes tags/flags, not paths.
complete -c skilldozer -f

# Flag matrix (§6.1/§6.2). --relative and --no-color have NO short forms.
complete -c skilldozer -s v -l version  -d 'Print the skilldozer version'
complete -c skilldozer -s h -l help     -d 'Show this help message'
complete -c skilldozer -s p -l path     -d 'Print the resolved skills directory'
complete -c skilldozer -s l -l list     -d 'Human-readable catalog (TAG, NAME, DESCRIPTION)'
complete -c skilldozer -s a -l all      -d 'Print every skill directory path, sorted by tag'
complete -c skilldozer -s f -l file     -d 'Print the SKILL.md path instead of the directory'
complete -c skilldozer       -l relative -d 'Print paths relative to the skills directory'
complete -c skilldozer       -l no-color -d 'Disable ANSI color'
# --search/-s take a free-text query, so NO completion is offered after them.
# We deliberately do NOT pass -r here: in fish 4.x `-r` switches into
# "complete the option's value" mode, which BYPASSES the global `-f` above and
# offers file names for the query. Without -r, --search/-s are treated as plain
# flags, so after `--search ` the global `-f` (no-files) applies and nothing is
# offered -- exactly the PRD §6.1 free-text-query behavior. (fish's -r is only a
# completion hint; skilldozer itself enforces that --search needs a value, exit 1.)
complete -c skilldozer -s s -l search -d 'Substring search over tag/name/description/keywords'

# --store <dir> (PRD §8.2): Non-interactive store path for init. Unlike --search,
# --store's value is a DIRECTORY, so here we DO pass `-r`: in fish 4.x `-r`
# switches into "complete the option's value" mode, which BYPASSES the global
# `-f` above and offers file/dir paths for the value. This is the intentional
# INVERSE of --search's no-`-r` (free-text -> offer nothing). No short form.
complete -c skilldozer -l store -d 'Non-interactive store path for init' -r

# `check` AND `init` are EXCLUSIVE subcommands (PRD §6.3). Offer them only as
# the first arg.
complete -c skilldozer -n '__fish_is_first_arg' -a 'check' -d 'Validate every skill on disk'
complete -c skilldozer -n '__fish_is_first_arg' -a 'init' -d 'First-run setup: pick/create the skills store and write the config'
complete -c skilldozer -n '__fish_is_first_arg' -a 'completion' -d 'Emit the shell completion script for eval'

# Dynamic tags: ONE directive with command substitution (NOT a hardcoded line per
# tag — the store is manifest-free and changes as skills are added). Suppressed
# once `check` OR `init` is seen (exclusive subcommand, PRD §6.3) AND when the
# previous arg is --search/-s (free-text query — no tag completion there either).
complete -c skilldozer -n 'not __fish_seen_subcommand_from check init completion; and not __fish_prev_arg_in --search -s' \
    -a '(skilldozer --relative --all 2>/dev/null)' -d 'skill tag'
