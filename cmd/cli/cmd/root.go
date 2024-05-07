package cmd

import (
	"os"

	"github.com/dennishilgert/apollo/cmd/cli/cmd/image"
	"github.com/dennishilgert/apollo/cmd/cli/cmd/messaging"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/spf13/cobra"
)

var log = logger.NewLogger("apollo.cli")

var rootCommand = &cobra.Command{
	Use:   "apollo",
	Short: "CLI for managing apollo",
	Long:  "Command Line Interface for managing the apollo FaaS platform",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

var logFlags = logger.ParseFlags()

func initFlags() {
	rootCommand.PersistentFlags().AddFlagSet(logFlags.FlagSet())
}

func init() {
	initFlags()

	rootCommand.AddCommand(image.Command)
	rootCommand.AddCommand(messaging.Command)
}

func Run() {
	if err := rootCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}
