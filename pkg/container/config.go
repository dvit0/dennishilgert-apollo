package container

import (
	"fmt"
	"os"
	"time"

	"github.com/dennishilgert/apollo/pkg/configuration"
	"github.com/spf13/viper"
)

type Config struct {
	// Name of the rootfs image file.
	RootFsImageFileName string

	// Specifies the directories whose content is excluded from export.
	RootFsExcludeDirs []string

	// Specifies the Docker container tag to use for the worker container.
	RootFsCopyTimeout time.Duration

	// Specifies the mount target for the destination directory on the host.
	BuilderContainerImageTag string

	// Specifies the mount target for the rootfs image.
	ContainerDestMountTarget string

	// Specifies the time the container is given to shutdown gracefully.
	ContainerImageMountTarget string

	// Specifies the comand that is run when image export container starts.
	ContainerStopTimeout int

	// Specifies the time the image export command is given to copy the rootfs.
	ContainerCommand []string

	// Username for the authentication with the image registry.
	ImageRegistryUsername string

	// Password for the authentication with the image registry.
	ImageRegistryPassword string
}

func DefaultConfig() Config {
	return Config{
		RootFsImageFileName:       "rootfs.ext4",
		RootFsExcludeDirs:         []string{"/boot", "/opt", "/proc", "/run", "/srv", "/sys", "/tmp"},
		RootFsCopyTimeout:         time.Duration(time.Second * 15),
		BuilderContainerImageTag:  "debian:stable",
		ContainerDestMountTarget:  "/dist",
		ContainerImageMountTarget: "/rootfs-image",
		ContainerStopTimeout:      5,
		ContainerCommand:          []string{"/bin/sh"},
		ImageRegistryUsername:     "",
		ImageRegistryPassword:     "",
	}
}

func LoadConfig() Config {
	var config Config

	// automatically load environment variables that match
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APOLLO")

	// loading the values from the environment or use default values
	configuration.LoadOrDefault("RootFsImageFileName", "APOLLO_ROOTFS_IMAGE_NAME", DefaultConfig().RootFsImageFileName)
	configuration.LoadOrDefault("BuilderContainerImageTag", "APOLLO_BUILDER_CONTAINER_IMAGE_TAG", DefaultConfig().BuilderContainerImageTag)
	configuration.LoadOrDefault("ContainerDestMountTarget", "APOLLO_CONTAINER_DEST_MOUNT_TARGET", DefaultConfig().ContainerDestMountTarget)
	configuration.LoadOrDefault("ContainerImageMountTarget", "APOLLO_CONTAINER_IMAGE_MOUNT_TARGET", DefaultConfig().ContainerImageMountTarget)

	// unmarshalling the Config struct
	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("unable to unmarshal logger config: %v\n", err)
		os.Exit(1)
	}

	return config
}
