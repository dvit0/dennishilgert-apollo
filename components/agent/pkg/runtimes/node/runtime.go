package node

import "apollo/agent/pkg/runtimes"

type NodeRuntime struct {
	runtimes.Runtime
}

func NewNodeRuntime() runtimes.DefaultRuntime {
	return &NodeRuntime{
		Runtime: runtimes.Runtime{
			Name: "node",
			Path: "/usr/bin/node",
		},
	}
}

func (r *NodeRuntime) Name() string {
	return r.Runtime.Name
}

func (r *NodeRuntime) Path() string {
	return r.Runtime.Path
}

func (r *NodeRuntime) Command() string {
	return r.Path()
}

func (r *NodeRuntime) CommandArgs() []string {
	return []string{"/workspace/node_wrapper.js"}
}
