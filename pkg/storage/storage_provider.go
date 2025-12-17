package storage

import (
	"io"
)

// StorageProvider defines how we save files (Local or Cloud)
type StorageProvider interface {
	Upload(filename string, content io.Reader) (string, error)
}