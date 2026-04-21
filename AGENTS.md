# AGENTS.md

## What this repo is

`github.com/rojbar/copia` — a CLI tool that renders Go `text/template` files from a "template package" directory (`.tmpl` files + `package_vars.json`) and writes output files. Single binary, no server.

---

## Commands

```bash
go build              # produces ./copia binary
go run . <pkg-path>   # run without building
go test ./...         # all tests
go test ./internal/parser/ -run TestFunctionName  # single test
gofmt -w .            # format (no Makefile/Taskfile)
```

No Makefile, Taskfile, CI workflows, or linter config exist — use raw `go` commands.

---

## Architecture

```
main.go               → cmd.Execute()
cmd/root.go           → wires parser + writer, runs Cobra root command
internal/parser/      → reads package_vars.json, walks dir, collects templates
internal/writer/      → renders text/template, strips .tmpl, writes files
```

- `writer` imports `parser` for the `TemplatePackage` type — no interface boundary between them.
- Functional options pattern on both `PackageParser` and `PackageWriter` (`WithMaxDepth`, `WithIgnorePatterns`, `WithVarsFileName`, `WithOutputDir`, `WithDryRun`, `WithVerbose`, `WithVarExtension`).
- Only dependency: `github.com/spf13/cobra v1.10.1`.

---

## Key conventions

- `package_vars.json` must be at the **root** of the template directory. Subdirectory vars files are not recognized (`isVariablesFile` matches the bare filename only, no path prefix).
- `.tmpl` extension is stripped from output filenames (`main.go.tmpl` → `main.go`). Non-`.tmpl` files are still processed as templates but keep their name.
- Template engine is `text/template` (not `html/template`). Variables referenced as `{{.keyName}}`.
- Sentinel errors: `parser.ErrMaxDepthExceeded`, `writer.ErrEmptyOutputDir` — check with `errors.Is`.
- CLI flags: `-o/--output` (default `./output`), `-d/--dry-run`, `-v/--verbose`.
