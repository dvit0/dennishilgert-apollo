package runtimes

type DefaultRuntime interface {
	Name() string
	Path() string
	Command() string
	CommandArgs() []string
}

type Runtime struct {
	Name string
	Path string
}
