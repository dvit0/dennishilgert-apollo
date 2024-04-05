package runtime

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/shared/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

var log = logger.NewLogger("apollo.agent.runtime")

type Config struct {
	BinaryPath string
	BinaryArgs []string
}

type Context struct {
	Runtime        string `json:"runtime"`
	RuntimeVersion string `json:"runtimeVersion"`
	RuntimeHandler string `json:"runtimeHandler"`
	MemoryLimit    int32  `json:"memoryLimit"`
	VCpuCores      int32  `json:"vCpuCores"`
}

type Event struct {
	EventUuid string      `json:"eventUuid"`
	EventType string      `json:"eventType"`
	Data      interface{} `json:"data"`
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
	EventUuid     string
	Status        int
	StatusMessage string
	Duration      string
	Logs          []LogLine
	Errors        []shared.Error
	Data          map[string]interface{}
}

type PersistentRuntime interface {
	Lock()
	Unlock()
	Initialize(handler string) error
	Invoke(ctx context.Context, fnCtx Context, fnEvt Event) (*Result, error)
	Close() error
}

type persistentRuntime struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr *bufio.Reader
	lock   sync.Mutex
}

func NewPersistentRuntime(ctx context.Context, config Config) (PersistentRuntime, error) {
	cmd := exec.CommandContext(ctx, config.BinaryPath, config.BinaryArgs...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &persistentRuntime{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
		stderr: bufio.NewReader(stderr),
	}, nil
}

func (p *persistentRuntime) Lock() {
	p.lock.Lock()
}

func (p *persistentRuntime) Unlock() {
	p.lock.Unlock()
}

func (p *persistentRuntime) Initialize(handler string) error {
	initBytes, err := json.Marshal(map[string]interface{}{
		"handler": handler,
	})
	if err != nil {
		log.Error("error while encoding init params to json")
		return err
	}
	_, err = p.stdin.Write(initBytes)
	if err != nil {
		log.Error("failed to write to stdin of the process")
		return err
	}
	_, err = p.stdin.Write([]byte("\n")) // Ensure the initialization message is completed
	if err != nil {
		log.Error("failed to write initialization end delimiter")
		return err
	}
	return nil
}

// Invoke invokes the user-provided code within the specified runtime.
func (p *persistentRuntime) Invoke(ctx context.Context, fnCtx Context, fnEvt Event) (*Result, error) {
	p.Lock()
	defer p.Unlock()

	timestamp := time.Now()

	fnParams := map[string]interface{}{
		"context": fnCtx,
		"event":   fnEvt,
	}
	fnParamsBytes, err := json.Marshal(fnParams)
	if err != nil {
		log.Error("error while encoding fn params to json")
		return nil, err
	}
	_, err = p.stdin.Write(fnParamsBytes)
	if err != nil {
		log.Error("failed to write to stdin of the process")
		return nil, err
	}
	_, err = p.stdin.Write([]byte("\n")) // Delimiter for end of request.
	if err != nil {
		log.Error("failed to write delimiter to stdin of the process")
		return nil, err
	}

	var errs []shared.Error
	var logs []LogLine
	var data map[string]interface{}

	outputDone := make(chan bool)
	go func() {
		for {
			line, _, err := p.stdout.ReadLine()
			if err != nil {
				if err != io.EOF {
					errs = append(errs, shared.Error{Message: "error while reading stdout", Cause: err.Error(), Stack: err.Error()})
					log.Errorf("error while reading stdout: %v", err)
				}
				break
			}
			var defaultProps DefaultProperties
			if err := json.Unmarshal(line, &defaultProps); err != nil {
				errs = append(errs, shared.Error{Message: "failed to decode json from output line", Cause: err.Error(), Stack: err.Error()})
				log.Errorf("failed to decode json from output line: %v", line)
				continue
			}

			switch defaultProps.Type {
			case "log":
				var logProps LogProperties
				if err := json.Unmarshal(defaultProps.RawProperties, &logProps); err != nil {
					errs = append(errs, shared.Error{Message: "failed to unmarshal log properties", Cause: err.Error(), Stack: err.Error()})
					log.Errorf("failed to unmarshal log properties: %v", err)

					logs = append(logs, LogLine{Timestamp: defaultProps.Timestamp, Level: "undefined", Message: string(line)})
					continue
				}
				logs = append(logs, LogLine{Timestamp: defaultProps.Timestamp, Level: logProps.Level, Message: logProps.Message})
			case "error":
				var errorProps ErrorProperties
				if err := json.Unmarshal(defaultProps.RawProperties, &errorProps); err != nil {
					errs = append(errs, shared.Error{Message: "failed to unmarshal error properties", Cause: err.Error(), Stack: err.Error()})
					log.Errorf("failed to unmarshal error properties: %v", err)

					logs = append(logs, LogLine{Timestamp: defaultProps.Timestamp, Level: "undefined", Message: string(line)})
					errs = append(errs, shared.Error{Message: string(line)})
					continue
				}
				logs = append(logs, LogLine{Timestamp: defaultProps.Timestamp, Level: "error", Message: errorProps.Message})
				errs = append(errs, shared.Error{Code: errorProps.Code, Message: errorProps.Message, Cause: errorProps.Cause, Stack: errorProps.Stack})
			case "result":
				var resultProps ResultProperties
				if err := json.Unmarshal(defaultProps.RawProperties, &resultProps); err != nil {
					errs = append(errs, shared.Error{Message: "failed to unmarshal result properties", Cause: err.Error(), Stack: err.Error()})
					log.Errorf("failed to unmarshal result properties: %v", err)
					continue
				}
				data = resultProps.Data
			default:
				log.Errorf("unsupported data type in output buffer: %v", defaultProps.Type)
			}
		}

		outputDone <- true
	}()

	<-outputDone // Wait for the output processing to complete

	// Check for errors on stderr
	errOutput, _ := io.ReadAll(p.stderr)
	if len(errOutput) > 0 {
		log.Errorf("runtime stderr: %s", string(errOutput))
	}

	duration := fmt.Sprintf("%dms", time.Since(timestamp).Milliseconds())

	status := 200
	statusMsg := "ok"
	if len(errs) > 0 {
		status = 500
		statusMsg = "error while invoking function"
	}

	return &Result{
		EventUuid:     fnEvt.EventUuid,
		Status:        status,
		StatusMessage: statusMsg,
		Duration:      duration,
		Logs:          logs,
		Errors:        errs,
		Data:          data,
	}, nil
}

func (p *persistentRuntime) Close() error {
	if err := p.stdin.Close(); err != nil {
		return err
	}
	return p.cmd.Wait()
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
