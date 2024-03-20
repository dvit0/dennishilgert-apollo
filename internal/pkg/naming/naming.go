package naming

import (
	"fmt"
)

func ImageNameRootFs(functionUuid string) string {
	return fmt.Sprintf("%s-rootfs", functionUuid)
}

func ImageNameCode(functionUuid string) string {
	return fmt.Sprintf("%s-code", functionUuid)
}

func ImagePrefix() string {
	return "apollo"
}

func ImageRefStr(imageRegistryAddress string, imageName string, imageTag string) string {
	return fmt.Sprintf("%s/%s/%s:%s", imageRegistryAddress, ImagePrefix(), imageName, imageTag)
}
