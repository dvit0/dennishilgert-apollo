package build

import (
	"apollo/cli/configs"
	"apollo/cli/pkg/container"
	"apollo/cli/pkg/utils"
	"context"
	"os"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "image-build",
	Short: "Build a Docker image from a Dockerfile",
	Long:  "",
	Run:   run,
}

var (
	logConfig     = configs.NewLoggingConfig()
	commandConfig = configs.NewImageBuildCommandConfig()
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

	rootLogger := logConfig.NewLogger("image-build")

	sourcePathStat, err := os.Stat(commandConfig.SourcePath)
	if err != nil {
		rootLogger.Error("error while resolving --source-path path", "reason", err)
		return 1
	}
	if !sourcePathStat.IsDir() {
		rootLogger.Error("value of --source-path does not point to a directory")
		return 1
	}

	dockerfileStat, err := os.Stat(commandConfig.Dockerfile)
	if err != nil {
		rootLogger.Error("error while resolving --dockerfile path", "reason", err)
		return 1
	}
	if dockerfileStat.IsDir() {
		rootLogger.Error("value of --dockerfile points to a directory")
		return 1
	}

	if commandConfig.ImageTag != "" {
		if !utils.IsValidTag(commandConfig.ImageTag) {
			rootLogger.Error("value of --image-tag is not a valid Docker image tag")
			return 1
		}
	}

	dockerClient, err := container.GetDefaultClient()
	if err != nil {
		rootLogger.Error("failed to get default Docker client", "reason", err)
		return 1
	}

	if err := container.ImageBuild(context.Background(), dockerClient, rootLogger, commandConfig.SourcePath, commandConfig.Dockerfile, commandConfig.ImageTag); err != nil {
		rootLogger.Error("failed to build Docker image", "reason", err)
		return 1
	}

	rootLogger.Info("image built successfully")

	return 0
}
