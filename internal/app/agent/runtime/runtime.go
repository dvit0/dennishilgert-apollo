package runtime

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/dennishilgert/apollo/internal/pkg/logger"
	logspb "github.com/dennishilgert/apollo/internal/pkg/proto/logs/v1"
	sharedpb "github.com/dennishilgert/apollo/internal/pkg/proto/shared/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var log = logger.NewLogger("apollo.agent.runtime")

type Config struct {
	BinaryPath  string
	BinaryArgs  []string
	Environment []string
}

type Context struct {
	Runtime        string `json:"runtime"`
	RuntimeVersion string `json:"runtimeVersion"`
	RuntimeHandler string `json:"runtimeHandler"`
	MemoryLimit    int32  `json:"memoryLimit"`
	VCpuCores      int32  `json:"vCpuCores"`
}

type Event struct {
	EventUuid string            `json:"eventUuid"`
	EventType string            `json:"eventType"`
	SourceIp  string            `json:"sourceIp"`
	Headers   map[string]string `json:"headers"`
	Params    map[string]string `json:"params"`
	Payload   interface{}       `json:"payload"`
}

type DefaultProperties struct {
	Timestamp     int64           `json:"timestamp"`
	Type          string          `json:"type"`
	RawProperties json.RawMessage `json:"properties"`
}

type Result struct {
	EventUuid     string
	Status        int
	StatusMessage string
	Duration      int64
	Logs          []*logspb.LogEntry
	Errors        []sharedpb.Error
	Data          map[string]interface{}
}

type PersistentRuntime interface {
	Config() Config
	Lock()
	Unlock()
	Start(handler string) error
	Close()
	Ready()
	Wait() error
	Tidy() error
	Invoke(ctx context.Context, event Event) (*Result, error)
}

type persistentRuntime struct {
	cfg     Config
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	lock    sync.Mutex
	readyCh chan bool
}

// NewPersistentRuntime creates a new PersistentRuntime instance.
func NewPersistentRuntime(ctx context.Context, config Config) (PersistentRuntime, error) {
	cmd := exec.CommandContext(ctx, config.BinaryPath, config.BinaryArgs...)
	cmd.Env = append(cmd.Env, config.Environment...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdin pipe failed: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdout pipe failed: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stderr pipe failed: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting command failed: %w", err)
	}

	return &persistentRuntime{
		cfg:     config,
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		readyCh: make(chan bool, 1),
	}, nil
}

// Config returns the runtime configuration.
func (p *persistentRuntime) Config() Config {
	return p.cfg
}

// Lock locks the runtime.
func (p *persistentRuntime) Lock() {
	p.lock.Lock()
}

// Unlock unlocks the runtime.
func (p *persistentRuntime) Unlock() {
	p.lock.Unlock()
}

// Start starts the runtime with the specified handler.
func (p *persistentRuntime) Start(handler string) error {
	initParamsBytes, err := json.Marshal(map[string]interface{}{
		"handler": handler,
	})
	if err != nil {
		return fmt.Errorf("encoding init params to json failed: %w", err)
	}
	_, err = p.stdin.Write(initParamsBytes)
	if err != nil {
		return fmt.Errorf("writing init params to stdin failed: %w", err)
	}
	_, err = p.stdin.Write([]byte("\n")) // Ensure the initialization message is completed
	if err != nil {
		return fmt.Errorf("writing end delimiter failed: %w", err)
	}
	return nil
}

// Close closes the ready channel.
func (p *persistentRuntime) Close() {
	close(p.readyCh)
}

// Ready signals that the runtime is ready.
func (p *persistentRuntime) Ready() {
	p.readyCh <- true
}

// Waits for the runtime to be ready and the command to finish.
func (p *persistentRuntime) Wait() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Wait for the runtime to be ready.
	select {
	case <-p.readyCh:
		// Runtime is ready, wait for the command to finish.
		return p.cmd.Wait()
	case <-ctx.Done():
		return errors.New("timeout while waiting for the runtime to be ready")
	}
}

// Tidy closes all open command pipes.
func (p *persistentRuntime) Tidy() error {
	if err := p.stdin.Close(); err != nil {
		return fmt.Errorf("closing stdin pipe failed: %w", err)
	}
	if err := p.stdout.Close(); err != nil {
		return fmt.Errorf("closing stdout pipe failed: %w", err)
	}
	if err := p.stderr.Close(); err != nil {
		return fmt.Errorf("closing stderr pipe failed: %w", err)
	}
	return nil
}

// Invoke invokes the user-provided code within the specified runtime.
func (p *persistentRuntime) Invoke(ctx context.Context, event Event) (*Result, error) {
	p.Lock()
	defer p.Unlock()

	var (
		errs  []sharedpb.Error
		logs  []*logspb.LogEntry
		data  map[string]interface{}
		start = time.Now()
	)

	if err := p.sendInvocationData(event); err != nil {
		return nil, fmt.Errorf("sending invocation data failed: %w", err)
	}

	if err := p.processOutput(&logs, &errs, &data); err != nil {
		return nil, fmt.Errorf("processing output failed: %w", err)
	}

	duration := time.Since(start)
	return p.buildResult(event.EventUuid, logs, errs, data, duration), nil
}

// sendInvocationData sends the invocation data to the runtime.
func (p *persistentRuntime) sendInvocationData(event Event) error {
	log.Debugf("sending invocation data to runtime for event: %s", event.EventUuid)
	fnParams := map[string]interface{}{"event": event}
	fnParamsBytes, err := json.Marshal(fnParams)
	if err != nil {
		return fmt.Errorf("encoding fn params to json failed: %w", err)
	}
	if _, err := p.stdin.Write(fnParamsBytes); err != nil {
		return fmt.Errorf("writing fn params to stdin failed: %w", err)
	}
	_, err = p.stdin.Write([]byte("\n"))
	return err
}

// processOutput reads the output buffer and processes the data.
func (p *persistentRuntime) processOutput(logs *[]*logspb.LogEntry, errs *[]sharedpb.Error, data *map[string]interface{}) error {
	log.Debug("processing output from runtime")
	reader := bufio.NewReader(p.stdout)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				return fmt.Errorf("reading from stdout failed: %w", err)
			}
			log.Warnf("error while reading from output buffer: %v", err)
			break
		}

		var defaultProps DefaultProperties
		if err := json.Unmarshal(line, &defaultProps); err != nil {
			return fmt.Errorf("decoding json from output line failed: %w", err)
		}

		if continueReading := p.handleLine(&defaultProps, logs, errs, data); !continueReading {
			return nil
		}
	}
	return nil
}

// handleLine processes a line from the output buffer.
func (p *persistentRuntime) handleLine(defaultProps *DefaultProperties, logs *[]*logspb.LogEntry, errs *[]sharedpb.Error, data *map[string]interface{}) bool {
	switch defaultProps.Type {
	case "log", "error":
		var props struct {
			Level   string `json:"level,omitempty"`
			Message string `json:"message"`
			Code    int32  `json:"code,omitempty"`
			Cause   string `json:"cause,omitempty"`
			Stack   string `json:"stack,omitempty"`
		}
		if err := json.Unmarshal(defaultProps.RawProperties, &props); err != nil {
			*errs = append(*errs, sharedpb.Error{Message: "failed to unmarshal properties", Cause: err.Error()})
			log.Errorf("failed to unmarshal properties: %v", err)
			return true
		}
		if defaultProps.Type == "log" {
			*logs = append(*logs, &logspb.LogEntry{Timestamp: &timestamppb.Timestamp{
				Seconds: defaultProps.Timestamp,
			}, LogLevel: props.Level, LogMessage: props.Message})
		} else {
			*errs = append(*errs, sharedpb.Error{Code: props.Code, Message: props.Message, Cause: props.Cause, Stack: props.Stack})
			*logs = append(*logs, &logspb.LogEntry{Timestamp: &timestamppb.Timestamp{
				Seconds: defaultProps.Timestamp,
			}, LogLevel: "error", LogMessage: props.Message})
		}
	case "result":
		var props struct {
			Data map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(defaultProps.RawProperties, &props); err != nil {
			*errs = append(*errs, sharedpb.Error{Message: "failed to unmarshal properties", Cause: err.Error()})
			log.Errorf("failed to unmarshal properties: %v", err)
			return true
		}
		*data = props.Data
	case "done":
		log.Debug("reading from stdout done")
		return false
	default:
		log.Errorf("unsupported data type in output buffer: %v", defaultProps.Type)
	}
	return true
}

// buildResult creates a Result instance from the provided data.
func (p *persistentRuntime) buildResult(eventUuid string, logs []*logspb.LogEntry, errs []sharedpb.Error, data map[string]interface{}, duration time.Duration) *Result {
	status, statusMessage := 200, "ok"
	if len(errs) > 0 {
		status, statusMessage = 500, "error while invoking function"
	}
	return &Result{
		EventUuid:     eventUuid,
		Status:        status,
		StatusMessage: statusMessage,
		Duration:      duration.Milliseconds(),
		Logs:          logs,
		Errors:        errs,
		Data:          data,
	}
}
