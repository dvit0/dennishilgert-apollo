package naming

import (
	"fmt"
)

var (
	// Prefix for the images.
	ImagePrefix = "apollo"

	MessagingFunctionInitializationTopic = "apollo_function_initialization"
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
