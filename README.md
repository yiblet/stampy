# Stampy

A command-line utility to prepend timestamps to lines of text, reading from stdin or a file and emitting to stdout or a file.

## Usage

```bash
stampy [OPTIONS] [INPUT_FILE] [OUTPUT_FILE]
```

### Options

- `-f, --format` - Timestamp format (default: `2006-01-02:15:04:05`)
  - Accepts Go time layouts
  - Special formats:
    - `s` - Elapsed time in seconds since start
    - `ms` - Elapsed time in milliseconds since start
    - `ds` - Time in milliseconds since last line
    - `dms` - Time in milliseconds since last line

### Arguments

- `INPUT_FILE` - Optional input file (defaults to stdin)
- `OUTPUT_FILE` - Optional output file (defaults to stdout)

## Examples

```bash
# Read from stdin, write to stdout with default timestamp format
echo "hello world" | stampy

# Process a file with elapsed seconds
stampy -f s input.txt

# Process stdin with milliseconds since last line, output to file
cat input.txt | stampy -f ds output.txt

# Process file to file with custom timestamp format
stampy -f "15:04:05" input.txt output.txt
```

## Output Format

Each line is prefixed with the timestamp followed by a colon and space:

```
2006-01-02:15:04:05: your text here
```