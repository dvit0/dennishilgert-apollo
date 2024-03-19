package build

import (
	"github.com/spf13/pflag"
)

type commandFlags struct {
	SourcePath           string
	Dockerfile           string
	ImageTag             string
	ImageRegistryAddress string
}

type parsedFlags struct {
	cmdFlags *commandFlags
	flagSet  *pflag.FlagSet
}

func ParseFlags() *parsedFlags {
	var f commandFlags

	fs := pflag.NewFlagSet("image-build", pflag.ExitOnError)
	fs.SortFlags = true

	// define default values only for reference, all flags are required
	fs.StringVar(&f.SourcePath, "source-path", ".", "Path to the source directory of the build context")
	fs.StringVar(&f.Dockerfile, "dockerfile", "./Dockerfile", "Path to the Dockerfile inside the build context")
	fs.StringVar(&f.ImageTag, "image-tag", "image:tag", "Tag to assign to the new Docker image")
	fs.StringVar(&f.ImageRegistryAddress, "image-registry-address", "host:port", "Network address of the image registry - optional with APOLLO_IMAGE_REGISTRY_ADDRESS set")

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
