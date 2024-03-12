package cmd

import (
	"os"

	"github.com/dennishilgert/apollo/cmd/cli/cmd/image"
	"github.com/dennishilgert/apollo/pkg/logger"
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

func init() {
	rootCommand.AddCommand(image.Command)
}

func Run() {
	if err := rootCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}
