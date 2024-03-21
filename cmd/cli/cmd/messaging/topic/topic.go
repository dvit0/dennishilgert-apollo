package topic

import (
	"os"

	"github.com/dennishilgert/apollo/cmd/cli/cmd/messaging/topic/create"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "topic",
	Short: "Manage messaging topics on the bootstrap servers",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

func init() {
	Command.AddCommand(create.Command)
}
