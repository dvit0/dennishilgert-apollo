package export

import (
	"context"
	"os"

	"github.com/dennishilgert/apollo/pkg/container"
	"github.com/dennishilgert/apollo/pkg/defers"
	"github.com/dennishilgert/apollo/pkg/logger"
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

	distPathStat, err := os.Stat(cmdFlags.CommandFlags().DistPath)
	if err != nil {
		log.Errorf("error while resolving --dist-path path: %v", err)
		return 1
	}
	if !distPathStat.IsDir() {
		log.Error("value of --dist-path does not point to a directory")
		return 1
	}

	dockerClient, err := container.GetDefaultClient()
	if err != nil {
		log.Errorf("failed to get default Docker client: %v", err)
		return 1
	}

	_, err = container.FetchImageIdByTag(context.Background(), dockerClient, log, cmdFlags.CommandFlags().ImageTag)
	if err != nil {
		log.Error("specified image does not exist")
		return 1
	}

	if err := container.ImageExport(context.Background(), dockerClient, log, cmdFlags.CommandFlags().DistPath, cmdFlags.CommandFlags().ImageTag); err != nil {
		log.Error("failed to export Docker image")
		return 1
	}

	log.Info("image exported successfully")

	return 0
}
