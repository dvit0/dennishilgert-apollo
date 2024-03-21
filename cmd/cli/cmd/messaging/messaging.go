package messaging

import (
	"os"

	"github.com/dennishilgert/apollo/cmd/cli/cmd/messaging/topic"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "messaging",
	Short: "Manage apollo messaging system",
	Long:  "Manage the messaging system for the apollo FaaS platform",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

func init() {
	Command.AddCommand(topic.Command)
}
