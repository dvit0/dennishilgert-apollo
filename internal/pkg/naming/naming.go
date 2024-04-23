package naming

import (
	"fmt"
	"os"
	"strconv"
)

var (
	// Prefix for the images.
	ImagePrefix = "apollo"

	// Messaging topics.
	MessagingFunctionInitializationTopic = "apollo_function_initialization"
	MessagingInstanceHeartbeatTopic      = "apollo_instance_heartbeat"

	// Name of the kernel storage bucket.
	StorageKernelBucketName = "apollo-kernels"

	// Name of the function storage bucket.
	StorageFunctionBucketName = "apollo-functions"
)

func MessagingWorkerRelatedAgentReadyTopic(workerUuid string) string {
	return fmt.Sprintf("apollo_agent_ready_related_%s", workerUuid)
}

func ImageNameRootFs(functionUuid string) string {
	return fmt.Sprintf("%s-rootfs", functionUuid)
}

func ImageNameCode(functionUuid string) string {
	return fmt.Sprintf("%s-code", functionUuid)
}

func ImageRefStr(imageRegistryAddress string, imageName string, imageTag string) string {
	return fmt.Sprintf("%s/%s/%s:%s", imageRegistryAddress, ImagePrefix, imageName, imageTag)
}

func FunctionStoragePath(dataPath string, functionUuid string) string {
	return fmt.Sprintf("%s/functions/%s", dataPath, functionUuid)
}

func FunctionImageFileName(functionUuid string) string {
	return fmt.Sprintf("%s.ext4", functionUuid)
}

func KernelStoragePath(dataPath string, kernelName string, kernelVersion string) string {
	return fmt.Sprintf("%s/kernels/%s-%s", dataPath, kernelName, kernelVersion)
}

func KernelFileName(kernelName string, kernelVersion string) string {
	return fmt.Sprintf("%s-%s", kernelName, kernelVersion)
}

func RuntimeStoragePath(dataPath string, runtimeName string, runtimeVersion string) string {
	return fmt.Sprintf("%s/runtimes/%s-%s", dataPath, runtimeName, runtimeVersion)
}

func RuntimeImageFileName(runtimeName string, runtimeVersion string) string {
	return fmt.Sprintf("%s-%s.ext4", runtimeName, runtimeVersion)
}

func RunnerStoragePath(dataPath string, runnerUuid string) string {
	return fmt.Sprintf("%s/runners/%s", dataPath, runnerUuid)
}

func RunnerLogFileName() string {
	return "firecracker.log"
}

func RunnerStdOutFileName() string {
	return "stdout.log"
}

func RunnerStdErrFileName() string {
	return "stderr.log"
}

func RunnerSocketFileName() string {
	return fmt.Sprintf("%s_fc.sock", strconv.Itoa(os.Getpid()))
}
