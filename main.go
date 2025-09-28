package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"

	"github.com/yiblet/stampy/internal/stampy"
)

type cliArgs struct {
	Template *string `arg:"positional" help:"Prefix template built from {elapsed}, {delta}, {time:<layout>}, {line}, and {}"`
	Input    string  `arg:"-i,--input" help:"Optional input file (defaults to stdin)"`
	Output   string  `arg:"-o,--output" help:"Optional output file (defaults to stdout)"`
	JSON     string  `arg:"--json" help:"Enable JSONL mode with specified timestamp key name"`
}

func (cliArgs) Description() string {
	return `Stampy prepends timestamps to each input line using a flexible brace template.

Template syntax (positional argument):
  - Omit the template to use the default "{iso}: {}".
  - Place {} where the original line should appear; if omitted the line is appended after the prefix.
  - Available tokens:
      {elapsed[:fmt]}  seconds since the first emitted line (fmt defaults to .1f)
      {delta[:fmt]}    seconds until the next line; final line always emits 0.0
      {time:<layout>}  absolute time using Go layouts (2006-01-02), Unix date directives (%Y-%m-%d), named layouts like iso/iso8601/iso8601nano, or the keyword unix
      {iso}           shortcut for RFC3339 (2006-01-02T15:04:05Z07:00)
      {unix[:fmt]}     seconds since the Unix epoch (default integer seconds)
      {line}           1-based line number
  - Escape literal braces with {{ and }}.

JSONL mode (--json <name>):
  - Outputs newline-delimited JSON objects instead of text
  - JSON objects get the timestamp merged in as {"<name>": "stamp", ...}
  - Primitives and arrays get wrapped as {"<name>": "stamp", "line": value}
  - Invalid JSON gets wrapped as {"<name>": "stamp", "line": "original"}

Examples:
  stampy                                  # default elapsed template
  stampy "{elapsed:.1f}s Δ{delta:.1f}s {}"  # elapsed + delta timings
  stampy "[{time:%H:%M:%S}] {line}: {}"     # human-readable clock with line numbers
  stampy --json ts "{elapsed:.2f}s"        # JSONL mode with "ts" timestamp key
`
}

func (c *cliArgs) toOptions() stampy.Options {
	opts := stampy.Options{
		Input:   c.Input,
		Output:  c.Output,
		JSONKey: c.JSON,
	}
	if c.Template != nil {
		opts.Template = *c.Template
		opts.TemplateProvided = true
	}
	return opts
}

func main() {
	var args cliArgs
	arg.MustParse(&args)
	if err := stampy.Run(args.toOptions()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
