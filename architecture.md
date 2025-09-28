# Stampy Architecture Overview

## High-Level Flow
1. CLI entrypoint parses flags and positional arguments, producing an `Options` struct.
2. The runtime wires input/output streams (`stdin`/`stdout` or files) and orchestrates line acquisition.
3. A `TemplateEngine` turns the prefix template (or default) into an evaluable plan.
4. The `LineProcessor` buffers lines, consults the clock, and computes `{elapsed}` and `{delta}` timestamps.
5. The `Emitter` serializes results either as formatted text (template mode) or enriched JSON objects (JSONL mode).

## Modules & Responsibilities
- **cmd/stampy (main.go)**: Owns CLI concerns only. Delegates to `internal/stampy` with a populated `Options`. Future flag changes (e.g. positional template, `--jsonl`) stay here.
- **internal/stampy/options.go (planned)**: Validates raw CLI input, filling defaults (template, delta semantics, jsonl key). Keeps `Options` clean for downstream components.
- **internal/stampy/io.go (existing createIO)**: Encapsulates reader/writer setup and cleanup management. Remains reusable for tests.
- **internal/stampy/template/**:
  - `parser.go`: Parses brace syntax into a slice of `Segment` nodes (literal text, insertion point, token with modifiers). Supports Go and `date(1)` layouts for `{time:...}` tokens.
  - `tokens.go`: Implements token evaluators. Each evaluator receives a `StampState` (elapsed, delta, line number, current time) and returns a string.
  - `compiler.go`: Resolves modifiers, validates combinations, and produces an executable `Template` (precomputed literal joins, function pointers, index of `{}`).
- **internal/stampy/clock.go**: Provides the clock interface and helpers (real clock vs. fake clock for tests).
- **internal/stampy/processor.go** (current implementation):
  - `lineBuffer` tracks pending lines, elapsed/delta timing, and line numbers in a testable unit.
  - `processLines` orchestrates reading, buffering, and delegating to emitters.
  - `StampState` (in template package) provides the render context per emission.
- **internal/stampy/emitter/**:
  - `text.go`: Applies the compiled template, inserts `{}` when requested, or appends the original line with a space when absent.
  - `jsonl.go`: Produces newline-delimited JSON objects. Attempts JSON decode; merges or wraps as required, preserving numeric/string/array types.

## Data Flow Details
```
stdin/file -> buffered reader -> LineProcessor (buffer queue)
      ↓                        ↓ uses clock timestamps
 TemplateEngine (compiled)     ↓ constructs stamp state {elapsed, delta, line, now}
      ↓                        ↓
 Emitter (text or JSONL) <- formatted stamp + original payload
      ↓
stdout/file
```

### Timing Mechanics
- Maintain `startTime`, `prevEmitTime`. For each new line timestamp `current`:
  - `elapsed = current - startTime`
  - `delta` for previous line = `current - prevEmitTime`
- Buffer the `(lineText, arrivalTime)` pair. When the next line arrives or EOF is reached, compute delta and emit.
- On EOF, flush the final buffer with `delta = 0`.

## Template Evaluation
- Segments are evaluated left-to-right:
  - `LiteralSegment`: appended verbatim.
  - `TokenSegment`: lookup via token registry (elapsed, delta, time, line).
  - `LineSegment`: placeholder for `{}`; replaced with original line text.
- If `{}` is omitted, `Emitter` concatenates template output, then a space, then line text.
- `time` token accepts Go layouts (`2006-01-02`), Unix `date` directives (`%Y-%m-%d`), and named presets like `iso`/`iso8601`/`iso8601nano`/`unix`; `iso` and `unix` are also exposed as dedicated tokens.
- Precision modifiers (e.g. `.1f`) rely on `fmt.Sprintf` verbs.
- Validation catches duplicated `{}` placeholders, unknown tokens, or mixing layout styles.

## JSONL Mode
- Triggered by `Options.JSONKey` being non-empty.
- Template still executes to produce the stamp string.
- Input line handling:
  - Parse using `encoding/json`.
  - If object: set `jsonObj[name] = stamp` (string), emit `json.Marshal`.
  - If primitive/array: wrap into `map[string]any{"name": stamp, "line": value}`.
  - If parse fails: wrap as `map[string]any{"name": stamp, "line": original}` where `original` is a string.
- Output uses compact encoding with deterministic key order (optional post-process for sorted keys if desired).

## Error Handling & Observability
- Parsing errors (template or JSON) return rich messages tagged with user input.
- IO errors propagate with `fmt.Errorf("read line: %w", err)` style wrapping.
- CLI reports errors to stderr with non-zero exit.
- Consider debug logging hooks guarded by an environment variable for future troubleshooting.

## Testing Strategy
- Unit tests per package (`template`, `processor`, `emitter`, `jsonl`).
- Table-driven tests covering brace parsing, time layout conversions, delta buffering, JSON wrapping, and error cases.
- Integration-style tests in `internal/stampy/stampy_test.go` to assert full pipeline behaviour, including JSONL mode with fake clocks.
