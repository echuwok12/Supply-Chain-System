package storage

import (
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	BaseDir string
}

func NewLocalStorage() *LocalStorage {
	// You can change this path to "C:/Users/Name/Documents/Storage" if you prefer
	// For now, we create a folder named "uploads" in the project root
	path := "uploads"

	// Ensure directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.Mkdir(path, os.ModePerm)
	}

	return &LocalStorage{BaseDir: path}
}

func (s *LocalStorage) Upload(filename string, content io.Reader) (string, error) {
	// Create the full path
	filePath := filepath.Join(s.BaseDir, filename)

	// Create the file on disk
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy data from memory to disk
	if _, err := io.Copy(dst, content); err != nil {
		return "", err
	}

	// Return the relative path (simulating a URL)
	return filePath, nil
}
