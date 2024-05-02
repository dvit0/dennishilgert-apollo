package export

import (
	"github.com/spf13/pflag"
)

type commandFlags struct {
	DistPath             string
	ImageTag             string
	ImageFileName        string
	ImageRegistryAddress string
	IncludeDirs          []string
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
	fs.StringVar(&f.ImageFileName, "image-file-name", "export.ext4", "Name of the file to export the image into")
	fs.StringVar(&f.ImageRegistryAddress, "image-registry-address", "host:port", "Network address of the image registry - optional with APOLLO_IMAGE_REGISTRY_ADDRESS set")
	fs.StringArrayVar(&f.IncludeDirs, "include-dir", []string{}, "Include directories in the image export (can be specified multiple times)")

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
