package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/shared/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

type Config struct {
	RuntimeBinaryPath string
	RuntimeBinaryArgs []string
}

type Context struct {
	Runtime        string `json:"runtime"`
	RuntimeVersion string `json:"runtimeVersion"`
	RuntimeHandler string `json:"runtimeHandler"`
	MemoryLimit    int32  `json:"memoryLimit"`
	VCpuCores      int32  `json:"vCpuCores"`
}

type Event struct {
	RequestID   string      `json:"requestId"`
	RequestType string      `json:"requestType"`
	Data        interface{} `json:"data"`
}

type DefaultProperties struct {
	Timestamp     int64           `json:"timestamp"`
	Type          string          `json:"type"`
	RawProperties json.RawMessage `json:"properties"`
}

type LogProperties struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

type ErrorProperties struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Cause   string `json:"cause"`
	Stack   string `json:"stack"`
}

type ResultProperties struct {
	Data map[string]interface{} `json:"data"`
}

type LogLine struct {
	Timestamp int64
	Level     string
	Message   string
}

type Result struct {
	RequestID     string
	Status        int
	StatusMessage string
	Duration      string
	Logs          []LogLine
	Error         *shared.Error
	Data          map[string]interface{}
}

// Invoke invokes the user-provided code within the specified runtime.
func Invoke(ctx context.Context, logger logger.Logger, fnCfg Config, fnCtx Context, fnEvt Event) (*Result, error) {
	fnParams := map[string]interface{}{
		"context": fnCtx,
		"event":   fnEvt,
	}
	fnParamsBytes, err := json.Marshal(fnParams)
	if err != nil {
		logger.Error("error while encoding fnParams to JSON")
		return nil, err
	}

	timestamp := time.Now()

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.CommandContext(ctx, fnCfg.RuntimeBinaryPath, fnCfg.RuntimeBinaryArgs...)
	cmd.Stdin = bytes.NewReader(fnParamsBytes)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	if err != nil {
		logger.Errorf("error while executing runtime: %s version %s", fnCtx.Runtime, fnCtx.RuntimeVersion)
		return nil, err
	}

	duration := fmt.Sprintf("%dms", time.Since(timestamp).Milliseconds())

	var data map[string]interface{}
	var fnErr *shared.Error
	var logs []LogLine

	decoder := json.NewDecoder(&stdoutBuf)
	for decoder.More() {
		var defaultProps DefaultProperties
		if err := decoder.Decode(&defaultProps); err != nil {
			if err == io.EOF {
				break // end of stream
			}
			logger.Errorf("failed to decode json from output buffer: %v", err)
			return nil, err
		}

		switch defaultProps.Type {
		case "log":
			var logProps LogProperties
			if err := json.Unmarshal(defaultProps.RawProperties, &logProps); err != nil {
				logger.Errorf("failed to unmarshal log properties: %v", err)
				return nil, err
			}
			logs = append(logs, LogLine{Timestamp: defaultProps.Timestamp, Level: logProps.Level, Message: logProps.Message})
		case "error":
			var errorProps ErrorProperties
			if err := json.Unmarshal(defaultProps.RawProperties, &errorProps); err != nil {
				logger.Errorf("failed to unmarshal error properties: %v", err)
				return nil, err
			}
			logs = append(logs, LogLine{Timestamp: defaultProps.Timestamp, Level: "error", Message: errorProps.Message})
			fnErr = &shared.Error{Code: errorProps.Code, Message: errorProps.Message, Cause: errorProps.Cause, Stack: errorProps.Stack}
		case "result":
			var resultProps ResultProperties
			if err := json.Unmarshal(defaultProps.RawProperties, &resultProps); err != nil {
				logger.Errorf("failed to unmarshal result properties: %v", err)
				return nil, err
			}
			data = resultProps.Data
		default:
			return nil, fmt.Errorf("unsupported data type in output buffer: %v", defaultProps.Type)
		}
	}

	status := 200
	statusMsg := "ok"
	if fnErr != nil {
		status = 500
		statusMsg = "error while invoking function"
	}

	return &Result{
		RequestID:     fnEvt.RequestID,
		Status:        status,
		StatusMessage: statusMsg,
		Duration:      duration,
		Logs:          logs,
		Error:         fnErr,
		Data:          data,
	}, nil
}

func LogsToStructList(logs []LogLine) ([]*structpb.Struct, error) {
	logList := make([]*structpb.Struct, 0, len(logs)) // Preallocate slice with the required capacity

	for _, logLine := range logs {
		structLine, err := structpb.NewStruct(map[string]interface{}{
			"timestamp": logLine.Timestamp,
			"level":     logLine.Level,
			"message":   logLine.Message,
		})
		if err != nil {
			return nil, err
		}
		logList = append(logList, structLine)
	}
	return logList, nil
}
