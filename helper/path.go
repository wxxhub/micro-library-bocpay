package helper

import (
	"os"
	"path/filepath"
)

var basePath string

func GetBasePath() string {
	if basePath == "" {
		workDir := os.Getenv("WORK_DIR")
		if workDir != "" {
			basePath = workDir
		} else {
			basePath, _ = filepath.Abs("")
		}
	}
	return basePath
}
