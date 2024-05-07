package push

import (
	"context"
	"fmt"
	"os"

	"github.com/dennishilgert/apollo/internal/pkg/container"
	"github.com/dennishilgert/apollo/internal/pkg/defers"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/utils"
	"github.com/spf13/cobra"
)

var log = logger.NewLogger("apollo.cli.image.push")

var Command = &cobra.Command{
	Use:   "push",
	Short: "Push a Docker image to the image registry",
	Long:  "",
	Run:   run,
}

var cmdFlags = ParseFlags()

func initFlags() {
	Command.Flags().AddFlagSet(cmdFlags.FlagSet())
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
	cleanup := defers.NewDefers()
	defer cleanup.CallAll()

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

	log.Infof("pushing image: %s", imageTag)
	if err := container.ImagePush(ctx, dockerClient, log, imageTag); err != nil {
		log.Fatalf("failed to push Docker image: %v", err)
	}

	log.Info("image pushed successfully")
	return 0
}
