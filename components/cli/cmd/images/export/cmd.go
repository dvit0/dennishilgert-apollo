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
	Use:   "image-export",
	Short: "Export a docker image as ext4 root filesystem",
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

	rootLogger := logConfig.NewLogger("image-export")

	dockerClient, err := containers.GetDefaultClient()
	if err != nil {
		rootLogger.Error("failed to get default Docker client", "reason", err)
		return 1
	}
	if err := containers.ImageExport(context.Background(), dockerClient, rootLogger, "/home/dennis/apollo-test/dist", "fc-rootfs:latest"); err != nil {
		rootLogger.Error("failed to export Docker image")
		return 1
	}

	rootLogger.Info("image exported successfully")

	return 0
}
