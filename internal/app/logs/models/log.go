package models

import logspb "github.com/dennishilgert/apollo/internal/pkg/proto/logs/v1"

type InvocationLogsEntry struct {
	FunctionUuid    string             `bson:"functionUuid"`
	FunctionVersion string             `bson:"functionVersion"`
	EventUuid       string             `bson:"eventUuid"`
	Logs            []*logspb.LogEntry `bson:"logs"`
}
