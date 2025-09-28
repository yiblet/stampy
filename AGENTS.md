# Repository Guidelines

## Project Structure & Module Organization
The CLI lives in `main.go`, which wires argument parsing, stream processing, and timestamp formatting. Unit tests reside in `main_test.go` and mirror the production functions; add new tests alongside the functions they exercise. Modules and dependencies are defined in `go.mod` and pinned in `go.sum`. User-facing docs and examples are kept in `README.md`. Keep additional assets in the project root unless a new directory is justified; prefer `internal/` for future packages.

## Build, Test, and Development Commands
Use `go run . --help` to confirm flag behaviour while iterating. `go build .` produces the `stampy` binary in the local directory. Run `go test ./...` before submitting changes; add `-cover` when validating coverage locally. When debugging file IO, `echo "demo" | go run . -f s` provides a quick smoke test.

## Coding Style & Naming Conventions
Follow standard Go formatting via `go fmt ./...`; indentation is tabs, and import order is managed by the formatter. Exported identifiers use PascalCase only when they must be public; otherwise prefer concise camelCase (`processLines`, `fakeClock`). Keep functions short and focused, factoring helpers into the same file until a new package is warranted. Surface errors with wrapped context via `fmt.Errorf("...: %w", err)` to match existing patterns.

## Testing Guidelines
Write table-driven tests when checking multiple scenarios; see `TestProcessLinesSecondsFormat` for shape. Fake time-sensitive behaviour with helper clocks instead of `time.Sleep`. Name tests `Test<FunctionScenario>` so `go test ./...` reports clearly. If a fix targets a regression, add a test first and ensure it fails before the change.

## Commit & Pull Request Guidelines
Commits in history follow Conventional Commit prefixes (`feat:`, `fix:`); continue that convention and keep messages imperative (`fix: handle empty input`). Each pull request should summarize behaviour changes, note test coverage (`go test ./...`), and link relevant issues. Include before/after samples when output formatting changes, and mention any manual verification steps for file IO.
