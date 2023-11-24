package helpers

import (
	"fmt"
	"path/filepath"

	"gocv.io/x/gocv"
)

func SaveMatToFile(mat gocv.Mat, dir string) error {
	// Create a complete file path
	filePath := filepath.Join(dir, "saved_image.png")

	// Use IMWrite to save the image
	if !gocv.IMWrite(filePath, mat) {
		return fmt.Errorf("failed to write image to file: %s", filePath)
	}

	return nil
}
