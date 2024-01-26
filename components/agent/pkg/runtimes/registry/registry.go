package registry

import (
	"apollo/agent/pkg/runtimes"
	"apollo/agent/pkg/runtimes/golang"
	"apollo/agent/pkg/runtimes/node"
)

func RuntimeByName(name string) runtimes.DefaultRuntime {
	runtimeList := []runtimes.DefaultRuntime{
		node.NewNodeRuntime(),
		golang.NewGolangRuntime(),
	}

	for _, runtime := range runtimeList {
		if runtime.Name() == name {
			return runtime
		}
	}
	return nil
}
