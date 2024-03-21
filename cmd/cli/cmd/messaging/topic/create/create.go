package create

import (
	"context"
	"os"

	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/messaging"
	"github.com/spf13/cobra"
)

var log = logger.NewLogger("apollo.cli.image.push")

var Command = &cobra.Command{
	Use:   "create",
	Short: "Create a messaging topic",
	Long:  "",
	Run:   run,
}

var cmdFlags = ParseFlags()

func initFlags() {
	Command.Flags().AddFlagSet(cmdFlags.FlagSet())
	Command.MarkFlagRequired("topic")
}

func init() {
	initFlags()
}

func run(cobraCommand *cobra.Command, args []string) {
	logger.ReadAndApply(cobraCommand, log)
	os.Exit(processCommand(cobraCommand.Context()))
}

func processCommand(ctx context.Context) int {
	var bootstrapServers string
	if cmdFlags.CommandFlags().BootstrapServers != "host:port" {
		bootstrapServers = cmdFlags.CommandFlags().BootstrapServers
	} else if os.Getenv("APOLLO_MESSAGING_BOOTSTRAP_SERVERS") != "" {
		bootstrapServers = os.Getenv("APOLLO_MESSAGING_BOOTSTRAP_SERVERS")
	}
	if bootstrapServers == "" {
		log.Fatalf("neither bootstrap servers flag nor env variable APOLLO_MESSAGING_BOOTSTRAP_SERVERS is set")
	}

	adminClient, err := messaging.GetDefaultAdminClient(bootstrapServers)
	if err != nil {
		log.Fatalf("failed to get default admin client: %v", err)
	}

	result, err := messaging.CreateTopic(ctx, adminClient, log, cmdFlags.CommandFlags().TopicName)
	if err != nil {
		log.Fatalf("failed to create topic: %s - reason: %v", cmdFlags.CommandFlags().TopicName, err)
	}

	log.Info("topic created successfully: %s", result.Topic)
	return 0
}
