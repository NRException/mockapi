package server

import (
	"errors"
	"fmt"
	"os"
)

func readFileContent(filePath string) (string, error) {
	b, err := os.ReadFile(filePath)

	if len(b) <= 0 {return "", fmt.Errorf("readFileContent: %w",errors.New("file has no content."))}
	if err != nil {return "", fmt.Errorf("readFileContent: %w",err)}
	
	return string(b), nil
}