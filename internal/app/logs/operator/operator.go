package operator

import (
	"context"
	"fmt"

	"github.com/dennishilgert/apollo/internal/app/logs/db"
	"github.com/dennishilgert/apollo/internal/app/logs/models"
	"github.com/dennishilgert/apollo/internal/pkg/logger"
	"github.com/dennishilgert/apollo/internal/pkg/naming"
	logspb "github.com/dennishilgert/apollo/internal/pkg/proto/logs/v1"
	messagespb "github.com/dennishilgert/apollo/internal/pkg/proto/messages/v1"
)

var log = logger.NewLogger("apollo.operator")

type LogsOperator interface {
	StoreInvocationLogs(ctx context.Context, message *messagespb.FunctionInvocationLogsMessage) error
	InvocationLogs(ctx context.Context, functionUuid, functionVersion string) (*logspb.FunctionInvocationLogsResponse, error)
}

type logsOperator struct {
	databaseClient db.DatabaseClient
}

func NewLogsOperator(databaseClient db.DatabaseClient) LogsOperator {
	return &logsOperator{
		databaseClient: databaseClient,
	}
}

func (l *logsOperator) StoreInvocationLogs(ctx context.Context, message *messagespb.FunctionInvocationLogsMessage) error {
	functionUuid, functionVersion, err := naming.ExtractFunctionDetailsFromIdentifier(message.FunctionIdentifier)
	if err != nil {
		return fmt.Errorf("failed to extract function details from identifier: %w", err)
	}

	log.Debugf("storing invocation logs for function %s:%s", functionUuid, functionVersion)
	return l.databaseClient.InsertLogs(ctx, models.InvocationLogsEntry{
		FunctionUuid:    functionUuid,
		FunctionVersion: functionVersion,
		EventUuid:       message.EventUuid,
		Logs:            message.Logs,
	})
}

func (l *logsOperator) InvocationLogs(ctx context.Context, functionUuid, functionVersion string) (*logspb.FunctionInvocationLogsResponse, error) {
	invocationLogEntries, err := l.databaseClient.GetLogs(ctx, functionUuid, functionVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get invocation logs: %w", err)
	}

	invocationLogs := make([]*logspb.InvocationLogs, 0, len(invocationLogEntries))
	for _, entry := range invocationLogEntries {
		invocationLogs = append(invocationLogs, &logspb.InvocationLogs{
			EventUuid: entry.EventUuid,
			Logs:      entry.Logs,
		})
	}

	return &logspb.FunctionInvocationLogsResponse{
		FunctionUuid:    functionUuid,
		FunctionVersion: functionVersion,
		InvocationLogs:  invocationLogs,
	}, nil
}
