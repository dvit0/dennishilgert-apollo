package configs

import "github.com/spf13/pflag"

type ImageBuildCommandConfig struct {
	flagBase

	SourcePath string
	Dockerfile string
	ImageTag   string
}

func NewImageBuildCommandConfig() *ImageBuildCommandConfig {
	return &ImageBuildCommandConfig{}
}

func (c *ImageBuildCommandConfig) FlagSet() *pflag.FlagSet {
	if c.initFlagSet() {
		c.flagSet.StringVar(&c.SourcePath, "source-path", ".", "Path to the source directory of the build context")
		c.flagSet.StringVar(&c.Dockerfile, "dockerfile", "./Dockerfile", "Path to the Dockerfile inside the build context")
		c.flagSet.StringVar(&c.ImageTag, "image-tag", "localhost:5000/apollo/baseos:bullseye", "Tag to assign to the new Docker image")
	}
	return c.flagSet
}

type ImageExportCommandConfig struct {
	flagBase

	DistPath string
	ImageTag string
}

func NewImageExportCommandConfig() *ImageExportCommandConfig {
	return &ImageExportCommandConfig{}
}

func (c *ImageExportCommandConfig) FlagSet() *pflag.FlagSet {
	if c.initFlagSet() {
		c.flagSet.StringVar(&c.DistPath, "dist-path", "/home/user/dist", "Abolsute path to the dist directory of the image export")
		c.flagSet.StringVar(&c.ImageTag, "image-tag", "localhost:5000/apollo/baseos:bullseye", "Tag of the Docker image to export")
	}
	return c.flagSet
}

type ImagePushCommandConfig struct {
	flagBase

	ImageTag string
}

func NewImagePushCommandConfig() *ImagePushCommandConfig {
	return &ImagePushCommandConfig{}
}

func (c *ImagePushCommandConfig) FlagSet() *pflag.FlagSet {
	if c.initFlagSet() {
		c.flagSet.StringVar(&c.ImageTag, "image-tag", "localhost:5000/apollo/baseos:bullseye", "Tag of the Docker image to push to the registry")
	}
	return c.flagSet
}
