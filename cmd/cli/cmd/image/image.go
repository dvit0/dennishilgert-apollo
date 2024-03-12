package image

import (
	"os"

	"github.com/dennishilgert/apollo/cmd/cli/cmd/image/build"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "image",
	Short: "Manage apollo function images",
	Long:  "Manage the function images for the apollo FaaS platform",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

func init() {
	Command.AddCommand(build.Command)
}
