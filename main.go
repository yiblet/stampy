package main

import (
	"fmt"
	"os"

	"github.com/alexflint/go-arg"

	"github.com/yiblet/stampy/internal/stampy"
)

type cliArgs struct {
	Format string `description:"Timestamp format; accepts Go layouts or 's'/'ms' for elapsed time or ds for time since last line and dms for time since last line (ms)" default:"2006-01-02:15:04:05" required:"true" short:"f" long:"format"`
	Input  string `description:"Optional input file (defaults to stdin)" arg:"positional"`
	Output string `description:"Optional output file (defaults to stdout)" arg:"positional"`
}

func (c *cliArgs) toOptions() stampy.Options {
	return stampy.Options{
		Format: c.Format,
		Input:  c.Input,
		Output: c.Output,
	}
}

func main() {
	var args cliArgs
	arg.MustParse(&args)
	if err := stampy.Run(args.toOptions()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
