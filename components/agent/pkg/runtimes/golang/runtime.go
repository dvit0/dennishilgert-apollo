package golang

import "apollo/agent/pkg/runtimes"

type GolangRuntime struct {
	runtimes.Runtime
}

func NewGolangRuntime() runtimes.DefaultRuntime {
	return &GolangRuntime{
		Runtime: runtimes.Runtime{
			Name: "golang",
			Path: "/usr/local/go/bin/go",
		},
	}
}

func (r *GolangRuntime) Name() string {
	return r.Runtime.Name
}

func (r *GolangRuntime) Path() string {
	return r.Runtime.Path
}

func (r *GolangRuntime) Command() string {
	return ""
}

func (r *GolangRuntime) CommandArgs() []string {
	return []string{}
}
