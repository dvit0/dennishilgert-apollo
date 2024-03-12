package flags

import (
	"github.com/spf13/pflag"
)

type FlagParser struct {
	flagSet *pflag.FlagSet
}

func NewFlagParser(name string) *FlagParser {
	fs := pflag.NewFlagSet(name, pflag.ExitOnError)
	fs.SortFlags = true

	return &FlagParser{
		flagSet: fs,
	}
}

func (f *FlagParser) FlagSet() *pflag.FlagSet {
	return f.flagSet
}

func (f *FlagParser) Parse(origArgs []string) {
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
	// ignore errors; pflag is set for ExitOnError
	_ = f.flagSet.Parse(args)
}
