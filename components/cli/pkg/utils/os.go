package utils

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"golang.org/x/sys/unix"
)

// Trap will listen for all stop signals and pass them along to the given process.
func Trap(process *os.Process) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		process.Signal(<-c)
	}()
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
