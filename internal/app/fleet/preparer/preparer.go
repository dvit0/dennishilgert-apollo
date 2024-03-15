package preparer

import (
	"fmt"
	"os"
	"strings"

	"github.com/dennishilgert/apollo/pkg/logger"
	"github.com/dennishilgert/apollo/pkg/proto/fleet/v1"
	"github.com/dennishilgert/apollo/pkg/utils"
)

var log = logger.NewLogger("apollo.manager.preparer")

type Options struct {
	DataPath string
}

type RunnerPreparer struct {
	dataPath string
}

func NewRunnerPreparer(opts Options) *RunnerPreparer {
	return &RunnerPreparer{
		dataPath: opts.DataPath,
	}
}

func (r *RunnerPreparer) PrepareDataDir() error {
	exists, fileInfo := utils.FileExists(r.dataPath)
	if !exists {
		if err := os.MkdirAll(r.dataPath, 0777); err != nil {
			return fmt.Errorf("failed to create data directory: %v", err)
		}
		return nil
	}
	ok, err := utils.IsDirAndWritable(r.dataPath, fileInfo)
	if !ok {
		return fmt.Errorf("data path is not a directory or not writable: %v", err)
	}
	return nil
}

func (r *RunnerPreparer) InitializeFunction(request *fleet.InitializeFunctionRequest) error {
	path := strings.Join([]string{r.dataPath, request.FunctionUuid}, string(os.PathSeparator))

	exists, fileInfo := utils.FileExists(path)
	if exists {
		ok, err := utils.IsDirAndWritable(r.dataPath, fileInfo)
		if !ok {
			return fmt.Errorf("target path is not a directory or not writable: %v", err)
		}
		empty, err := utils.IsDirEmpty(path)
		if !empty {
			return fmt.Errorf("target directory is not empty")
		}
		if err != nil {
			return fmt.Errorf("failed to check if target directory is empty: %v", err)
		}
		return nil
	}
	if err := os.MkdirAll(r.dataPath, 0777); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// TODO: download assets into created directory
	return nil
}
