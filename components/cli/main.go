package main

import (
	"fmt"
	"os"

	build "apollo/cli/cmd/images/build"
	export "apollo/cli/cmd/images/export"
	push "apollo/cli/cmd/images/push"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:   "apollocli",
	Short: "acli",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

func init() {
	rootCommand.AddCommand(build.Command)
	rootCommand.AddCommand(push.Command)
	rootCommand.AddCommand(export.Command)
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}
