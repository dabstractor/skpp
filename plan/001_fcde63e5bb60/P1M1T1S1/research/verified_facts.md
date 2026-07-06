# Verified facts — P1.M1.T1.S1 (go.mod, go.sum, .gitignore, LICENSE)

All below confirmed by direct execution in the target environment
(Go 1.26.4, linux/amd64) on 2026-07-06.

## 1. `go mod init` writes the FULL toolchain version

```
$ go mod init github.com/dabstractor/skpp
$ cat go.mod
module github.com/dabstractor/skpp

go 1.26.4        # <-- full version, NOT 1.25
```

The contract requires `go 1.25` (latest-two-stable policy; 1.25 + 1.26 are the
two latest stable lines; toolchain is 1.26.4 so 1.25 is safe as a *minimum*).
=> The implementer MUST edit the directive from `go 1.26.4` to `go 1.25`
   (major.minor only, no patch). Do this AFTER `go mod init`, BEFORE `go get`.

## 2. `go get gopkg.in/yaml.v3` → v3.0.1 (latest), marked `// indirect`

```
$ go get gopkg.in/yaml.v3
go: added gopkg.in/yaml.v3 v3.0.1

$ cat go.mod
module github.com/dabstractor/skpp

go 1.26.4
require gopkg.in/yaml.v3 v3.0.1 // indirect
```

The `// indirect` comment is EXPECTED and CORRECT for this subtask: no .go file
imports yaml.v3 yet, so the module graph records it as indirect. It will be
promoted to a direct require once `internal/discover` imports it (later
subtask P1.M2.T4). Do NOT run `go mod tidy` to "fix" it — tidy would REMOVE
the entry entirely because nothing imports it yet, leaving no go.sum hash. Keep
the `// indirect` require line as-is.

## 3. `go build ./...` on an empty module succeeds (exit 0)

```
$ go build ./...
go: warning: "./..." matched no packages
$ echo $?
0
```

The warning "matched no packages" is HARMLESS and expected (no .go files yet).
Exit code is 0 = success. The success criterion "`go build ./...` succeeds" is
met. Do NOT silence the warning or add a dummy .go file to do so — this subtask
explicitly must NOT create main.go or any internal/ package.

## 4. go.sum contents (yaml.v3 v3.0.1)

```
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/yaml.v3 v3.0.1 h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
```

`go get` produces go.sum automatically. check.v1 is yaml.v3's test-only dep
(its hash is recorded but it is NOT downloaded into the build). This is normal.

## 5. CRITICAL GOTCHA — pre-existing .gitignore has WRONG content

The repo root ALREADY contains an untracked `.gitignore` (214 bytes) created by
the planning tooling, with GENERIC content that does NOT match PRD §16:

    dist/
    build/
    node_modules/
    vendor/
    venv/
    .env
    .env.*
    !.env.example
    .DS_Store
    Thumbs.db
    .pi-subagents/

This is WRONG for skpp. PRD §16 (authoritative) requires EXACTLY:

    /skpp
    /dist
    *.test
    *.out
    .DS_Store

=> The implementer MUST OVERWRITE the existing .gitignore with exactly the
   PRD §16 content. Do NOT append. Do NOT keep node_modules/.env/venv lines
   (this is a Go project). A naive "does .gitignore exist?" check would wrongly
   report success — verify the CONTENT.

## 6. Go module path & license

- Module path: `github.com/dabstractor/skpp` (matches git remote
  `git@github.com:dabstractor/skpp.git` and PRD §5).
- License: MIT (PRD §19 decision 11; mcpeepants has none but PRD requires it).
- Copyright holder for LICENSE: pull from `git config user.name` / `user.email`.
