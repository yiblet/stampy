package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"

	"github.com/yiblet/stampy/internal/stampy"
)

type cliArgs struct {
	Template *string `arg:"positional" optional:"true" description:"Prefix template built from {elapsed}, {delta}, {time:<layout>}, {line}, and {}"`
	Input    string  `help:"Optional input file (defaults to stdin)" short:"i" long:"input"`
	Output   string  `help:"Optional output file (defaults to stdout)" short:"o" long:"output"`
}

func (cliArgs) Description() string {
	return `Stampy prepends timestamps to each input line using a flexible brace template.

Template syntax (positional argument):
  - Omit the template to use the default "{elapsed:.1f}s {}".
  - Place {} where the original line should appear; if omitted the line is appended after the prefix.
  - Available tokens:
      {elapsed[:fmt]}  seconds since the first emitted line (fmt defaults to .1f)
      {delta[:fmt]}    seconds until the next line; final line always emits 0.0
      {time:<layout>}  absolute time using Go layouts (2006-01-02), Unix date directives (%Y-%m-%d), named layouts like iso/iso8601/iso8601nano, or the keyword unix
      {iso}           shortcut for RFC3339 (2006-01-02T15:04:05Z07:00)
      {unix[:fmt]}     seconds since the Unix epoch (default integer seconds)
      {line}           1-based line number
  - Escape literal braces with {{ and }}.

Examples:
  stampy                                  # default elapsed template
  stampy "{elapsed:.1f}s Δ{delta:.1f}s {}"  # elapsed + delta timings
  stampy "[{time:%H:%M:%S}] {line}: {}"     # human-readable clock with line numbers
`
}

func (c *cliArgs) toOptions() stampy.Options {
	opts := stampy.Options{
		Input:  c.Input,
		Output: c.Output,
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
