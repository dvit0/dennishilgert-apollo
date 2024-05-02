package export

import (
	"context"
	"fmt"
	"os"

	"github.com/dennishilgert/apollo/pkg/container"
	"github.com/dennishilgert/apollo/pkg/defers"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/utils"
	"github.com/spf13/cobra"
)

var log = logger.NewLogger("apollo.cli.image.export")

var Command = &cobra.Command{
	Use:   "export",
	Short: "Export a Docker image as ext4 root filesystem",
	Long:  "",
	Run:   run,
}

var cmdFlags = ParseFlags()

func initFlags() {
	Command.Flags().AddFlagSet(cmdFlags.FlagSet())
	Command.MarkFlagRequired("dist-path")
	Command.MarkFlagRequired("image-tag")
	Command.MarkFlagRequired("image-file-name")
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

	distPathStat, err := os.Stat(cmdFlags.CommandFlags().DistPath)
	if err != nil {
		log.Fatalf("error while resolving dist path: %v", err)
	}
	if !distPathStat.IsDir() {
		log.Fatalf("dist path does not point to a directory: %s", cmdFlags.CommandFlags().DistPath)
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

	_, err = container.FetchImageIdByTag(ctx, dockerClient, log, imageTag)
	if err != nil {
		log.Fatalf("image does not exist: %s", imageTag)
	}

	log.Infof("exporting image: %s", imageTag)
	if err := container.ImageExport(ctx, dockerClient, log, cmdFlags.CommandFlags().DistPath, imageTag, cmdFlags.CommandFlags().ImageFileName, cmdFlags.CommandFlags().IncludeDirs); err != nil {
		log.Fatalf("failed to export Docker image: %v", err)
	}

	log.Info("image exported successfully")
	return 0
}
