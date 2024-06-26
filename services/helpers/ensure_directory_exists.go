package helpers

import "os"

// ensures a specific directory exists, creating it if necessary
func EnsureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755) // 0755 means everyone can read, owner can write
	}
	return nil
}
