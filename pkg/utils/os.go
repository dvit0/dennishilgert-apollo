package utils

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"

	"golang.org/x/sys/unix"
)

type OsArch int32

const (
	Unknown OsArch = iota
	Arch_x86_64
	Arch_Arm_64
	Other
)

func DetectArchitecture() OsArch {
	var utsname unix.Utsname
	if err := unix.Uname(&utsname); err != nil {
		return Unknown
	}
	machine := string(utsname.Machine[:])
	switch machine {
	case "x86_64":
		return Arch_x86_64
	case "arm64", "aarch64":
		return Arch_Arm_64
	}
	return Other
}

func (o OsArch) String() string {
	return [...]string{"unknown", "x86_64", "arm64", "other"}[o]
}

// FileExists returns if the given path exists.
func FileExists(filePath string) fs.FileInfo {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil
	}
	return stat
}

// IsDir returns if the given file is directory.
func IsDir(fileInfo fs.FileInfo) bool {
	return fileInfo.Mode()&fs.ModeDir != 0
}

// IsSocket returns if the given file is a unix socket.
func IsSocket(fileInfo fs.FileInfo) bool {
	return fileInfo.Mode()&fs.ModeSocket != 0
}

// IsDirAndWritable checks if the given path exists, is a directory and is writable.
// Important: This function uses the unix package, which only works on unix systems
func IsDirAndWritable(path string) error {
	pathStat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !pathStat.IsDir() {
		return fmt.Errorf("given file is not a directory")
	}
	if err := unix.Access(path, unix.W_OK); err != nil {
		return err
	}
	return nil
}

// IsValidDirName checks if the given name is a valid dir name.
func IsValidDirName(name string) bool {
	pattern := `^(\/?)[a-zA-Z0-9_.][a-zA-Z0-9_.-]{0,254}$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched
}
