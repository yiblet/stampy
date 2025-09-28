# Stampy Prefix & JSONL Upgrade Requirements

## Expanded User Stories & Usage Sketches
- Incident response with elapsed + delta timings
  - Command: `kubectl logs deploy/api | stampy "{elapsed:.1f}s Δ{delta:.1f}s {}"`
  - Sample output:
    ```
    1.2s Δ0.0s starting probe
    3.4s Δ2.2s upstream call failed (502)
    3.4s Δ0.0s shutting down
    ```
- Batch replay with absolute wall clock and job identifiers
  - Command: `stampy "[{time:2006-01-02 15:04:05}] JOB42 {}" < events.log`
  - Sample output:
    ```
    [2024-08-04 09:00:00] JOB42 BEGIN stage=extract
    [2024-08-04 09:00:07] JOB42 records=5000
    ```
- Research experiments with human-friendly clock and annotations
  - Command: `python run_experiment.py | stampy "{time:%H:%M:%S} [trialA] {}"`
  - Sample output:
    ```
    14:03:10 [trialA] stimulus-on
    14:03:12 [trialA] response=0.87
    ```
- CLI pipelines needing selective prefixing for multiline records
  - Command: `tail -f build.log | stampy "#{line} {elapsed:.1f}s {}" --mode lines`
  - Sample output:
    ```
    #1 0.0s compiling service
    #2 1.5s tests running
    #3 0.0s tests passed
    ```
- Accessibility-oriented transcripts that require elapsed timing without losing raw text
  - Command: `recording | stampy "{elapsed:.1f}s {}"`
  - Sample output:
    ```
    12.3s Introduce yourself.
    15.8s Thanks for having me.
    ```
- Structured observability via JSONL enrichment
  - Command: `tail -f service.jsonl | stampy "{elapsed:.2f}" --jsonl elapsed_at`
  - Sample output:
    ```jsonl
    {"elapsed_at":"0.0","level":"info","msg":"startup"}
    {"elapsed_at":"2.1","level":"warn","msg":"retrying"}
    {"elapsed_at":"0.0","line":{"unexpected":true}}
    ```

## Formatting Language Goals
- Accept an optional positional template argument using single braces `{token[:modifier]}`; surrounding text is literal.
- Reserve bare `{}` as the insertion point for the original line body; if users omit `{}`, stampy appends the formatted prefix plus a space automatically.
- Core tokens:
  - `elapsed` → elapsed time since the first emitted line; precision modifiers follow `fmt` verbs (default `.1f`).
  - `delta` → time until the next line arrives; the final line emits `0.0` seconds by definition.
  - `time:<layout>` → absolute timestamp; accept Go layouts (`2006-01-02`), `date(1)` specifiers (`%Y-%m-%d`), or named layouts such as `iso`, `iso8601`, `iso8601nano`, and `unix`.
  - `iso` → shortcut for RFC3339 output.
  - `unix[:fmt]` → seconds since the Unix epoch (default integer seconds).
  - `line` → running line number starting at 1.
- Allow inline token mixing (e.g. `{time:%H:%M:%S} Δ{delta:.2f}s`); escape literal braces via doubling (`{{` / `}}`).
- Provide a sensible built-in template when no argument is supplied (e.g. `{elapsed:.1f}s {}`) so stampy remains useful out of the box.
- Emit descriptive parse errors for unknown tokens, malformed braces, mixed layout styles, or invalid precision values to keep feedback actionable.

## Delta Semantics and Processing Model
- Buffer each line until the subsequent line is read so that `{delta}` reflects time-to-next-line; emit the buffered line once the next line arrives.
- On EOF, flush the final buffered line immediately with `{delta}` forced to `0.0` seconds.
- Treat blank lines as events with their own elapsed/delta values, preserving pipeline fidelity.
- Document the contrast between `{elapsed}` (time since first line) and `{delta}` (time until next line) so users pick the right prefix.

## JSONL Enrichment Mode
- Introduce an optional `--jsonl [name]` flag; default key name is `"time"` when the flag is present without an argument.
- Evaluate the positional template to obtain the string stamp value (fall back to the default template when none is provided).
- For each input line:
  - Attempt JSON decoding.
  - If the line is a JSON object, inject or overwrite the key `name` with the stamp string and re-emit compact JSON.
  - If the line parses as an array/number/string/bool/null, wrap it as `{name: <stamp>, "line": <parsed_value>}`.
  - If parsing fails (raw text), emit `{name: <stamp>, "line": <original_text>}` with `line` as a string.
- Preserve delta buffering semantics so stamp calculation remains consistent even in JSON mode.
- Ensure JSON output is newline-delimited, stable-key ordered, and free of trailing whitespace for easy downstream consumption.

## Migration Considerations
- Drop legacy `-f` formats in favor of the positional brace template; show a concise error directing users to supply templates like `stampy "{elapsed:.0f} {}"` for their old use cases.
- Keep README examples, CLI help, and tests aligned with the positional argument model, highlighting real-world templates from the scenarios above.
- Consider shipping a `stampy templates list/show` helper once the base syntax is stable, but defer until the core language lands.
- Add regression tests for JSONL mode covering object enrichment, primitive wrapping, malformed JSON, and delta timing to guarantee consistent behaviour.
