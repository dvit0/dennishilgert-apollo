package utils

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"

	"golang.org/x/sys/unix"
)

type OsArch int32

const (
	Arch_Unknown OsArch = iota
	Arch_x86_64
	Arch_Arm_64
	Arch_Other
)

// DetectArchitecture detects the architecture of the system.
func DetectArchitecture() OsArch {
	var utsname unix.Utsname
	if err := unix.Uname(&utsname); err != nil {
		return Arch_Unknown
	}
	machine := strings.Trim(string(utsname.Machine[:]), "\x00")
	switch machine {
	case "x86_64":
		return Arch_x86_64
	case "arm64", "aarch64":
		return Arch_Arm_64
	}
	return Arch_Other
}

// String returns the string representation of the OsArch.
func (o OsArch) String() string {
	return [...]string{"unknown", "x86_64", "arm64", "other"}[o]
}

// FileExists returns if the given path exists.
func FileExists(filePath string) (bool, fs.FileInfo) {
	stat, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, stat
}

// IsDir returns if the given file is directory.
func IsDir(fileInfo fs.FileInfo) bool {
	return fileInfo.Mode()&fs.ModeDir != 0
}

// IsSocket returns if the given file is a unix socket.
func IsSocket(fileInfo fs.FileInfo) bool {
	return fileInfo.Mode()&fs.ModeSocket != 0
}

// IsWritable checks if the directory at the given path is writable.
// Important: This function uses the unix package, which only works on unix systems.
func IsWritable(path string) (bool, error) {
	if err := unix.Access(path, unix.W_OK); err != nil {
		return false, err
	}
	return true, nil
}

// IsDirEmpty checks if a given directory is empty.
func IsDirEmpty(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()
	_, err = file.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// IsDirAndWritable checks if the file is a directory and is writable.
func IsDirAndWritable(filePath string, fileInfo fs.FileInfo) (bool, error) {
	if dir := IsDir(fileInfo); !dir {
		return false, fmt.Errorf("given file path is not a directory: %s", filePath)
	}
	_, err := IsWritable(filePath)
	if err != nil {
		return false, fmt.Errorf("given file path is not writable: %v", err)
	}
	return true, nil
}

// IsValidDirName checks if the given name is a valid dir name.
func IsValidDirName(name string) bool {
	pattern := `^(\/?)[a-zA-Z0-9_.][a-zA-Z0-9_.-]{0,254}$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched
}
