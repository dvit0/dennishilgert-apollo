package create

import "github.com/spf13/pflag"

type commandFlags struct {
	BootstrapServers string
	TopicName        string
}

type parsedFlags struct {
	cmdFlags *commandFlags
	flagSet  *pflag.FlagSet
}

func ParseFlags() *parsedFlags {
	var f commandFlags

	fs := pflag.NewFlagSet("topic-create", pflag.ExitOnError)
	fs.SortFlags = true

	// define default values only for reference, all flags are required
	fs.StringVar(&f.BootstrapServers, "bootstrap-servers", "host:port", "Kafka bootstrap servers to connect to - optional with APOLLO_MESSAGING_BOOTSTRAP_SERVERS set")
	fs.StringVar(&f.TopicName, "topic", "apollo_function_initialization", "Name of the topic to create")

	return &parsedFlags{
		cmdFlags: &f,
		flagSet:  fs,
	}
}

func (p *parsedFlags) CommandFlags() *commandFlags {
	return p.cmdFlags
}

func (p *parsedFlags) FlagSet() *pflag.FlagSet {
	return p.flagSet
}
