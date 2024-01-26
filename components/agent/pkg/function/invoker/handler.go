package invoker

import (
	"apollo/agent/pkg/function"
	"apollo/agent/pkg/runtimes/registry"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/go-hclog"
)

func Invoke(ctx context.Context, opLogger hclog.Logger, fnCfg function.Config, fnContext function.Context, fnEvent function.Event) (interface{}, []map[string]interface{}, error) {
	input := map[string]interface{}{
		"context": fnContext,
		"event":   fnEvent,
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, nil, err
	}

	runtime := registry.RuntimeByName(fnCfg.Runtime)
	if runtime == nil {
		return nil, nil, fmt.Errorf("cannot find runtime %s", fnCfg.Runtime)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.CommandContext(ctx, runtime.Command(), runtime.CommandArgs()...)
	cmd.Stdin = bytes.NewReader(inputBytes)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()

	if err != nil {
		opLogger.Error("error while executing runtime", "reason", err)
		return nil, nil, err
	}

	var result interface{}
	var logs []map[string]interface{}

	scanner := bufio.NewScanner(&stdoutBuf)
	for scanner.Scan() {
		var output map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &output); err != nil {
			opLogger.Error("failed to parse json from output buffer", "reason", err)
			return nil, nil, err
		}
		switch output["type"] {
		case "log":
			logs = append(logs, map[string]interface{}{"timestamp": output["timestamp"], "level": strings.ToUpper(output["level"].(string)), "message": output["message"]})
		case "error":
			logs = append(logs, map[string]interface{}{"timestamp": output["timestamp"], "level": "ERROR", "code": output["code"], "message": output["message"], "cause": output["cause"], "stack": output["stack"]})
			opLogger.Error("error thrown while executing function code", "error", err)
			return result, logs, err
		case "result":
			result = output["data"]
		default:
			return result, logs, errors.New("unsupported data type in output buffer: " + output["type"].(string))
		}
	}
	if err := scanner.Err(); err != nil {
		opLogger.Error("error while scanning output buffer", "reason", err)
		return result, logs, err
	}

	return result, logs, nil
}
