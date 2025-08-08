package universal

import (
	"io/ioutil"
	"os"
)

// ReadFile reads content from a file
func ReadFile(filepath string) ([]byte, error) {
	return ioutil.ReadFile(filepath)
}

// FileExists checks if a file exists
func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}