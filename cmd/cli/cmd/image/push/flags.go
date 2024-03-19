package push

import (
	"github.com/spf13/pflag"
)

type commandFlags struct {
	ImageTag             string
	ImageRegistryAddress string
}

type parsedFlags struct {
	cmdFlags *commandFlags
	flagSet  *pflag.FlagSet
}

func ParseFlags() *parsedFlags {
	var f commandFlags

	fs := pflag.NewFlagSet("image-push", pflag.ExitOnError)
	fs.SortFlags = true

	// define default values only for reference, all flags are required
	fs.StringVar(&f.ImageTag, "image-tag", "localhost:5000/apollo/baseos:bullseye", "Tag of the Docker image to push to the registry")
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
