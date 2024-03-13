package build

import (
	"context"
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
	os.Exit(processCommand())
}

func processCommand() int {
	sourcePathStat, err := os.Stat(cmdFlags.CommandFlags().SourcePath)
	if err != nil {
		log.Errorf("error while resolving --source-path path: %v", err)
		return 1
	}
	if !sourcePathStat.IsDir() {
		log.Error("value of --source-path does not point to a directory")
		return 1
	}

	dockerfileStat, err := os.Stat(cmdFlags.CommandFlags().Dockerfile)
	if err != nil {
		log.Errorf("error while resolving --dockerfile path: %v", err)
		return 1
	}
	if dockerfileStat.IsDir() {
		log.Error("value of --dockerfile points to a directory")
		return 1
	}

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

	if err := container.ImageBuild(context.Background(), dockerClient, log, cmdFlags.CommandFlags().SourcePath, cmdFlags.CommandFlags().Dockerfile, cmdFlags.CommandFlags().ImageTag); err != nil {
		log.Errorf("failed to build Docker image: %v", err)
		return 1
	}

	log.Info("image built successfully")

	return 0
}
