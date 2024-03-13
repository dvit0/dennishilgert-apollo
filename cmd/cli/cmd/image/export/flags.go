package export

import (
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/spf13/pflag"
)

type commandFlags struct {
	DistPath string
	ImageTag string
	Logger   logger.Config
}

type parsedFlags struct {
	cmdFlags *commandFlags
	flagSet  *pflag.FlagSet
}

func ParseFlags() *parsedFlags {
	var f commandFlags

	fs := pflag.NewFlagSet("image-export", pflag.ExitOnError)
	fs.SortFlags = true

	// define default values only for reference, all flags are required
	fs.StringVar(&f.DistPath, "dist-path", "/home/user/dist", "Abolsute path to the dist directory of the image export")
	fs.StringVar(&f.ImageTag, "image-tag", "localhost:5000/apollo/baseos:bullseye", "Tag of the Docker image to export")

	f.Logger = logger.DefaultConfig()

	return &parsedFlags{
		cmdFlags: &f,
		flagSet:  fs,
	}
}

func (p *parsedFlags) CommandFlags() *commandFlags {
	return p.cmdFlags
}

func (p *parsedFlags) FlagSet() *pflag.FlagSet {
	return p.flagSet
}
