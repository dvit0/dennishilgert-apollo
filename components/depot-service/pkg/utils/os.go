package utils

import (
	"fmt"
	"io/fs"
	"os"
)

func CheckIfFileExistsAndIsRegular(filePath string) (fs.FileInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if !fileInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("file is not regular: %v", filePath)
	}
	return fileInfo, nil
}
