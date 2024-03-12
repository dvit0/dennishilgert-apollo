package build

import (
	"context"
	"os"

	"github.com/dennishilgert/apollo/pkg/container"
	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/utils"
	"github.com/spf13/cobra"
)

var log = logger.NewLogger("cli")

var Command = &cobra.Command{
	Use:   "build",
	Short: "Build a Docker image from a Dockerfile",
	Long:  "",
	Run:   run,
}

func run(cobraCommand *cobra.Command, args []string) {
	os.Exit(processCommand(args))
}

func processCommand(args []string) int {
	flags := New(args)

	if err := logger.ApplyConfigToLoggers(&flags.Logger); err != nil {
		log.Fatal(err)
	}

	sourcePathStat, err := os.Stat(flags.SourcePath)
	if err != nil {
		log.Errorf("error while resolving --source-path path: %v", err)
		return 1
	}
	if !sourcePathStat.IsDir() {
		log.Error("value of --source-path does not point to a directory")
		return 1
	}

	dockerfileStat, err := os.Stat(flags.Dockerfile)
	if err != nil {
		log.Errorf("error while resolving --dockerfile path: %v", err)
		return 1
	}
	if dockerfileStat.IsDir() {
		log.Error("value of --dockerfile points to a directory")
		return 1
	}

	if flags.ImageTag != "" {
		if !utils.IsValidTag(flags.ImageTag) {
			log.Error("value of --image-tag is not a valid Docker image tag")
			return 1
		}
	}

	dockerClient, err := container.GetDefaultClient()
	if err != nil {
		log.Errorf("failed to get default Docker client: %v", err)
		return 1
	}

	if err := container.ImageBuild(context.Background(), dockerClient, log, flags.SourcePath, flags.Dockerfile, flags.ImageTag); err != nil {
		log.Errorf("failed to build Docker image: %v", err)
		return 1
	}

	log.Info("image built successfully")

	return 0
}
