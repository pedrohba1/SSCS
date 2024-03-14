package helpers

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"gocv.io/x/gocv"
)

func SaveMatToFile(mat gocv.Mat, dir string) (string, error) {
	// Create a complete file path

	fname := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10) + ".jpg"

	filePath := filepath.Join(dir, fname)

	// Use IMWrite to save the image
	if !gocv.IMWrite(filePath, mat) {
		return "", fmt.Errorf("failed to write image to file: %s", filePath)
	}

	return filePath, nil
}
