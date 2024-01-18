package push

import (
	"apollo/cli/configs"
	"apollo/cli/pkg/containers"
	"apollo/cli/pkg/utils"
	"context"
	"os"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "image-push",
	Short: "Push a Docker image to the image registry",
	Long:  "",
	Run:   run,
}

var (
	logConfig     = configs.NewLoggingConfig()
	commandConfig = configs.NewImagePushCommandConfig()
)

func initFlags() {
	Command.Flags().AddFlagSet(logConfig.FlagSet())
	Command.Flags().AddFlagSet(commandConfig.FlagSet())
}

func init() {
	initFlags()
}

func run(cobraCommand *cobra.Command, _ []string) {
	os.Exit(processCommand())
}

func processCommand() int {
	cleanup := utils.NewDefers()
	defer cleanup.CallAll()

	rootLogger := logConfig.NewLogger("image-push")

	if commandConfig.ImageTag != "" {
		if !utils.IsValidTag(commandConfig.ImageTag) {
			rootLogger.Error("value of --image-tag is not a valid Docker image tag")
			return 1
		}
	}

	dockerClient, err := containers.GetDefaultClient()
	if err != nil {
		rootLogger.Error("failed to get default Docker client", "reason", err)
		return 1
	}

	if err := containers.ImagePush(context.Background(), dockerClient, rootLogger, commandConfig.ImageTag); err != nil {
		rootLogger.Error("failed to push Docker image", "reason", err)
		return 1
	}

	rootLogger.Info("image pushed successfully")

	return 0
}
