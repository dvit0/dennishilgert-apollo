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
	Short: "Export a Docker image as ext4 root filesystem",
	Long:  "",
	Run:   run,
}

var (
	logConfig     = configs.NewLoggingConfig()
	commandConfig = configs.NewImageExportCommandConfig()
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

	rootLogger := logConfig.NewLogger("image-export")

	distPathStat, err := os.Stat(commandConfig.DistPath)
	if err != nil {
		rootLogger.Error("error while resolving --dist-path path", "reason", err)
		return 1
	}
	if !distPathStat.IsDir() {
		rootLogger.Error("value of --dist-path does not point to a directory")
		return 1
	}

	dockerClient, err := containers.GetDefaultClient()
	if err != nil {
		rootLogger.Error("failed to get default Docker client", "reason", err)
		return 1
	}

	_, err = containers.FetchImageIdByTag(context.Background(), dockerClient, rootLogger, commandConfig.ImageTag)
	if err != nil {
		rootLogger.Error("specified image does not exist ")
		return 1
	}

	if err := containers.ImageExport(context.Background(), dockerClient, rootLogger, commandConfig.DistPath, commandConfig.ImageTag); err != nil {
		rootLogger.Error("failed to export Docker image")
		return 1
	}

	rootLogger.Info("image exported successfully")

	return 0
}
