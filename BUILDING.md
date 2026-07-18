# Building Language Betawi from source

This is the developer/build guide. For the project overview, language
syntax, and installation instructions, see [README.md](README.md).


## First-time setup

Fetch every dependency this project has accumulated (the language runtime
needs SQLite; the installer needs a Windows GUI toolkit, Windows syscall
bindings, and a resource-embedding tool):

```
go get modernc.org/sqlite
go get github.com/lxn/walk
go get github.com/lxn/win
go get golang.org/x/sys/windows/registry
go install github.com/akavel/rsrc@latest
go mod tidy
```

`rsrc` is a command-line tool, not a library — `go install` puts its binary
in `%USERPROFILE%\go\bin` (or wherever `go env GOPATH`\bin points). If
`build.bat` can't find it afterward, make sure that folder is on your PATH.

**Why admin rights are required:** the installer writes to the machine-wide
System PATH (`HKEY_LOCAL_MACHINE`), not just your personal user PATH — that
registry location is protected and needs elevation to write to. The
`.exe`'s embedded manifest declares `requireAdministrator`, so Windows
shows the UAC consent prompt automatically whether you double-click it or
launch it from an unelevated terminal.

## The installer's wizard flow

`betawi-installer.exe` is a standard multi-page setup wizard (Welcome →
Existing Installation choice → Progress → Finish), the same general shape
as installers like PostgreSQL's, using native Windows common controls
throughout — no custom-drawn graphics:

1. **Welcome** page.
2. **Existing installation** page — only shown if a previous `betawi.exe`
   is already at the target install path — offering Overwrite or Repair.
3. **Progress** page — a real `ProgressBar` control whose value is driven
   directly by bytes-written-so-far ÷ total bytes of the embedded compiler
   binary (chunked extraction, not a timed animation).
4. **Finish** page — shows "Betawi language successfully downloaded" on
   success, or the Betawi-phrased error text on failure.

## Building

**Just the language compiler** (cross-platform, run `.bwi` scripts):

```
go build -o betawi.exe .\cmd\betawi
.\betawi.exe examples\hello.bwi
```

**The full installer** (Windows-only — bundles betawi.exe + shows the
splash screen + registers PATH). This is a **two-stage build**: the
installer embeds the compiler binary via `go:embed`, so the compiler must
be built first. Just run:

```
build.bat
```

This produces two files:
- `betawi.exe` — the compiler itself
- `betawi-installer.exe` — the setup wizard end users should download

## Project layout

```
cmd/betawi/              the language compiler CLI
cmd/betawi-installer/    the Windows setup wizard (Windows-only)
internal/lexer/          tokenizer + Betawi synonym/fuzzy matching
internal/ast/            AST node definitions
internal/parser/         Pratt parser
internal/object/         runtime value types
internal/evaluator/      tree-walking interpreter + builtins (print/DB/HTTP)
internal/betawimsg/      shared Betawi-voice error/message templates
internal/installer/      wizard UI + setup flow (Windows-only)
internal/assets/         embedded compiler binary
examples/                sample .bwi scripts
```
