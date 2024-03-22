package naming

import (
	"fmt"
)

var (
	// Prefix for the images.
	ImagePrefix = "apollo"

	// Messaging topics
	MessagingFunctionInitializationTopic = "apollo_function_initialization"
	MessagingFunctionStatusUpdateTopic   = "apollo_function_status_update"

	// Name of the kernel storage bucket
	StorageKernelBucketName = "apollo-kernels"

	// Name of the function storage bucket
	StorageFunctionBucketName = "apollo-functions"
)

func ImageNameRootFs(functionUuid string) string {
	return fmt.Sprintf("%s-rootfs", functionUuid)
}

func ImageNameCode(functionUuid string) string {
	return fmt.Sprintf("%s-code", functionUuid)
}

func ImageRefStr(imageRegistryAddress string, imageName string, imageTag string) string {
	return fmt.Sprintf("%s/%s/%s:%s", imageRegistryAddress, ImagePrefix, imageName, imageTag)
}
