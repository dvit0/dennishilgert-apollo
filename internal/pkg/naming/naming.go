package naming

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	// Prefix for the images.
	ImagePrefix = "apollo"

	// Messaging topics.
	MessagingFunctionInitializationResponsesTopic   = "apollo_function_initialization_responses"
	MessagingFunctionDeinitializationRequestsTopic  = "apollo_function_deinitialization_requests"
	MessagingFunctionDeinitializationResponsesTopic = "apollo_function_deinitialization_responses"
	MessagingFunctionCodeUploadedTopic              = "apollo_function_code_uploaded"
	MessagingFunctionStatusUpdateTopic              = "apollo_function_status_update"
	MessagingInstanceHeartbeatTopic                 = "apollo_instance_heartbeat"

	// Names of the storage buckets.
	StorageKernelBucketName       = "apollo-kernels"
	StorageFunctionBucketName     = "apollo-functions"
	StorageDependenciesBucketName = "apollo-dependencies"
)

func CacheAddLeaseDeclaration(key string) string {
	return fmt.Sprintf("lease:%s", key)
}

func CacheStripLeaseDeclaration(key string) string {
	stripped, _ := strings.CutPrefix(key, "lease:")
	return stripped
}

func CacheIsLeaseKey(key string) bool {
	return strings.HasPrefix(key, "lease:")
}

func CacheServiceInstanceKeyName(instanceUuid string) string {
	return fmt.Sprintf("service:%s", instanceUuid)
}

func CacheIsServiceInstanceLease(key string) bool {
	return strings.Split(key, ":")[0] == "service"
}

func CacheWorkerInstanceKeyName(workerUuid string) string {
	return fmt.Sprintf("worker:%s", workerUuid)
}

func CacheIsWorkerInstanceLease(key string) bool {
	return strings.Split(key, ":")[0] == "worker"
}

func CacheArchitectureSetKey(architecture string) string {
	return fmt.Sprintf("arch:%s", architecture)
}

func CacheFunctionSetKey(functionIdentifier string) string {
	return fmt.Sprintf("function:%s", functionIdentifier)
}

func CacheExtractInstanceUuid(key string) string {
	parts := strings.Split(key, ":")
	return parts[len(parts)-1]
}

func CacheExtractServiceInstanceType(key string) string {
	return strings.Split(key, ":")[1]
}

func MessagingWorkerRelatedAgentReadyTopic(workerUuid string) string {
	return fmt.Sprintf("apollo_agent_ready_related_%s", workerUuid)
}

func ImageRefStr(imageRegistryAddress string, imageName string, imageTag string) string {
	return fmt.Sprintf("%s/%s/%s:%s", imageRegistryAddress, ImagePrefix, imageName, imageTag)
}

func FunctionIdentifier(functionUuid string, functionVersion string) string {
	return fmt.Sprintf("%s_%s", functionUuid, functionVersion)
}

func FunctionStoragePath(dataPath string, functionUuid string) string {
	return fmt.Sprintf("%s/functions/%s", dataPath, functionUuid)
}

func FunctionStoragePathBase(dataPath string) string {
	return fmt.Sprintf("%s/functions", dataPath)
}

func FunctionImageFileName(functionVersion string) string {
	return fmt.Sprintf("%s.ext4", functionVersion)
}

func FunctionExtractVersionFromImageFileName(imageFileName string) string {
	return strings.Split(imageFileName, ".")[0]
}

func FunctionCodeStorageName(functionUuid string, functionVersion string) string {
	return fmt.Sprintf("%s/%s.zip", functionUuid, functionVersion)
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

func RuntimeInitialCodeDeclarator() string {
	return "initial"
}

func RuntimeInitialImageRefStr(imageRegistryAddress string, runtimeName string, runtimeVersion string, runtimeArchitecture string) string {
	return ImageRefStr(imageRegistryAddress, fmt.Sprintf("initial-%s", runtimeName), fmt.Sprintf("%s-%s", runtimeVersion, runtimeArchitecture))
}
