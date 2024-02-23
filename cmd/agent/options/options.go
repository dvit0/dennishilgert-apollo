package options

import (
	"flag"

	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/spf13/pflag"
)

type Options struct {
	ApiPort int
	Logger  logger.Options
}

func New(origArgs []string) *Options {
	var opts Options

	// We are using pflag to parse the CLI flags
	// pflag is a drop-in replacement for the standard library's "flag" package, however…
	// There's one key difference: with the stdlib's "flag" package, there are no short-hand options so options can be defined with a single slash.
	// With pflag, single slashes are reserved for shorthands.
	// So, we are doing this thing where we iterate through all args and double-up the slash if it's single
	// This works *as long as* we don't start using shorthand flags (which haven't been in use so far).
	args := make([]string, len(origArgs))
	for i, a := range origArgs {
		if len(a) > 2 && a[0] == '-' && a[1] != '-' {
			args[i] = "-" + a
		} else {
			args[i] = a
		}
	}

	// create a flag set
	fs := pflag.NewFlagSet("agent", pflag.ExitOnError)
	fs.SortFlags = true

	fs.IntVar(&opts.ApiPort, "api-port", 50051, "the port used for the api server")

	opts.Logger = logger.DefaultOptions()
	opts.Logger.AttachCmdFlags(flag.StringVar, flag.BoolVar)

	// ignore errors; pflag is set for ExitOnError
	_ = fs.Parse(args)

	return &opts
}
