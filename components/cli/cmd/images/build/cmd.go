package build

import (
	"apollo/cli/configs"
	"apollo/cli/pkg/containers"
	"apollo/cli/pkg/utils"
	"context"
	"os"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "image-build",
	Short: "Build a docker image from a Dockerfile",
	Long:  "",
	Run:   run,
}

var (
	logConfig = configs.NewLoggingConfig()
)

func initFlags() {
	Command.Flags().AddFlagSet(logConfig.FlagSet())
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

	rootLogger := logConfig.NewLogger("image-build")

	dockerClient, err := containers.GetDefaultClient()
	if err != nil {
		rootLogger.Error("failed to get default Docker client", "reason", err)
		return 1
	}
	if err := containers.ImageBuild(context.Background(), dockerClient, rootLogger, ".", "./Dockerfile.rootfs", "fc-rootfs:latest"); err != nil {
		rootLogger.Error("failed to build Docker image", "reason", err)
		return 1
	}

	rootLogger.Info("image built successfully")

	return 0
}
