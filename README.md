# Stampy

A command-line utility to prepend timestamps to lines of text, reading from stdin or a file and emitting to stdout or a file. Stampy now supports a brace-based template language for building expressive prefixes.

## Usage

```bash
stampy [TEMPLATE] [--input PATH] [--output PATH]
```

### Template Basics

- Omit the template to use the default `"{elapsed:.1f}s {}"` (elapsed seconds plus the original line).
- Use `{}` to choose where the original line is inserted; if omitted, stampy appends the line after the rendered prefix with a space.
- Available tokens:
  - `{elapsed[:fmt]}` – seconds since the first line (default `:.1f`).
  - `{delta[:fmt]}` – seconds until the next line; the final line always shows `0.0`.
  - `{time:<layout>}` – absolute timestamp using Go layouts (`2006-01-02`), Unix `date` directives (`%Y-%m-%d`), or named layouts such as `iso`, `iso8601`, `iso8601nano`, and `unix`.
  - `{iso}` – shortcut for RFC3339 (`2006-01-02T15:04:05Z07:00`).
  - `{unix[:fmt]}` – seconds since the Unix epoch (default integer seconds).
  - `{line}` – 1-based line number.
- Escape literal braces with `{{` or `}}`.

### Options

- `--input, -i` – optional input file (defaults to stdin).
- `--output, -o` – optional output file (defaults to stdout).

## Examples

```bash
# Read from stdin, write to stdout with default elapsed template
echo "hello world" | stampy

# Show elapsed and delta timings
tail -f app.log | stampy "{elapsed:.1f}s Δ{delta:.1f}s {}"

# Human-readable clock time for files
stampy "[{time:15:04:05}] {}" --input input.txt --output output.txt

# Unix timestamp output
stampy "{unix} {}" --input input.txt

# Line numbers with elapsed time
cat script.log | stampy "#{line} {elapsed:.1f}s {}"
```

## Output Format

Each line is prefixed according to the chosen template. For the default template:

```
0.0s your text here
```
