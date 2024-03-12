package build

import (
	"flag"

	"github.com/dennishilgert/apollo/pkg/flags"
	"github.com/dennishilgert/apollo/pkg/logger"
)

type cmdFlags struct {
	SourcePath string
	Dockerfile string
	ImageTag   string

	Logger logger.Config
}

func New(origArgs []string) *cmdFlags {
	var f cmdFlags

	parser := flags.NewFlagParser("image-build")
	parser.FlagSet().StringVar(&f.SourcePath, "source-path", ".", "Path to the source directory of the build context")
	parser.FlagSet().StringVar(&f.Dockerfile, "dockerfile", "./Dockerfile", "Path to the Dockerfile inside the build context")
	parser.FlagSet().StringVar(&f.ImageTag, "image-tag", "localhost:5000/apollo/baseos:bullseye", "Tag to assign to the new Docker image")

	f.Logger = logger.DefaultConfig()
	f.Logger.AttachCmdFlags(flag.StringVar, flag.BoolVar)

	parser.Parse(origArgs)

	return &f
}
