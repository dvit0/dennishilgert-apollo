package push

import (
	"context"
	"os"

	"github.com/dennishilgert/apollo/pkg/container"
	"github.com/dennishilgert/apollo/pkg/defers"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/utils"
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
	os.Exit(processCommand())
}

func processCommand() int {
	if err := logger.ApplyConfigToLoggers(&cmdFlags.CommandFlags().Logger); err != nil {
		log.Fatal(err)
	}

	cleanup := defers.NewDefers()
	defer cleanup.CallAll()

	if cmdFlags.CommandFlags().ImageTag != "" {
		if !utils.IsValidTag(cmdFlags.CommandFlags().ImageTag) {
			log.Error("value of --image-tag is not a valid Docker image tag")
			return 1
		}
	}

	dockerClient, err := container.GetDefaultClient()
	if err != nil {
		log.Errorf("failed to get default Docker client: %v", err)
		return 1
	}

	if err := container.ImagePush(context.Background(), dockerClient, log, cmdFlags.CommandFlags().ImageTag); err != nil {
		log.Errorf("failed to push Docker image: %v", err)
		return 1
	}

	log.Info("image pushed successfully")

	return 0
}
