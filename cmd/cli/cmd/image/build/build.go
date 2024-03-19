package build

import (
	"context"
	"fmt"
	"os"

	"github.com/dennishilgert/apollo/pkg/container"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/utils"
	"github.com/spf13/cobra"
)

var log = logger.NewLogger("apollo.cli.image.build")

var Command = &cobra.Command{
	Use:   "build",
	Short: "Build a Docker image from a Dockerfile",
	Long:  "",
	Run:   run,
}

var cmdFlags = ParseFlags()

func initFlags() {
	Command.Flags().AddFlagSet(cmdFlags.FlagSet())
	Command.MarkFlagRequired("source-path")
	Command.MarkFlagRequired("dockerfile")
	Command.MarkFlagRequired("image-tag")
}

func init() {
	initFlags()
}

func run(cobraCommand *cobra.Command, args []string) {
	logger.ReadAndApply(cobraCommand, log)
	os.Exit(processCommand(cobraCommand.Context()))
}

func processCommand(ctx context.Context) int {
	sourcePathStat, err := os.Stat(cmdFlags.CommandFlags().SourcePath)
	if err != nil {
		log.Fatalf("error while resolving source path: %v", err)
	}
	if !sourcePathStat.IsDir() {
		log.Fatalf("source path does not point to a directory: %s", cmdFlags.CommandFlags().SourcePath)
	}

	dockerfileStat, err := os.Stat(cmdFlags.CommandFlags().Dockerfile)
	if err != nil {
		log.Fatalf("error while resolving dockerfile path: %v", err)
	}
	if dockerfileStat.IsDir() {
		log.Fatalf("dockerfile points to a directory: %s", cmdFlags.CommandFlags().Dockerfile)
	}

	if cmdFlags.CommandFlags().ImageTag != "" {
		if !utils.IsValidTag(cmdFlags.CommandFlags().ImageTag) {
			log.Fatalf("given image tag is invalid: %s", cmdFlags.CommandFlags().ImageTag)
		}
	}

	var imageRegistryAddress string
	if cmdFlags.CommandFlags().ImageRegistryAddress != "host:port" {
		imageRegistryAddress = cmdFlags.CommandFlags().ImageRegistryAddress
	} else if os.Getenv("APOLLO_IMAGE_REGISTRY_ADDRESS") != "" {
		imageRegistryAddress = os.Getenv("APOLLO_IMAGE_REGISTRY_ADDRESS")
	}
	if imageRegistryAddress == "" {
		log.Fatalf("neither image registry address flag nor env variable APOLLO_IMAGE_REGISTRY_ADDRESS is set")
	}
	if !utils.IsValidRegistryAddress(imageRegistryAddress) {
		log.Fatalf("image registry address is invalid")
	}

	dockerClient, err := container.GetDefaultClient()
	if err != nil {
		log.Fatalf("failed to get default Docker client: %v", err)
	}

	imageTag := fmt.Sprintf("%s/apollo/%s", imageRegistryAddress, cmdFlags.CommandFlags().ImageTag)

	log.Infof("building image: %s", imageTag)
	if err := container.ImageBuild(
		ctx, dockerClient,
		log,
		cmdFlags.CommandFlags().SourcePath,
		cmdFlags.CommandFlags().Dockerfile,
		imageTag,
	); err != nil {
		log.Fatalf("failed to build Docker image: %v", err)
	}

	log.Info("image built successfully")
	return 0
}
