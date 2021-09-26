package storage

import (
	"fmt"
	"io"
	"sync"

	"github.com/rs/zerolog/log"
)

// MemoryStore an implementation of the object storage
// that writes to files to an in memory hashmap
type MemoryStore struct {
	objects sync.Map
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// ReadObject reads a file from the local storage directory to the provided io.Writer
func (s *MemoryStore) ReadObject(id string, w io.Writer) error {
	// load the object from the hashmap
	obj, ok := s.objects.Load(id)
	if !ok {
		return ErrFileDoesNotExist
	}

	data, ok := obj.([]byte)
	if !ok {
		return ErrFileDoesNotExist
	}

	// write the data to the requester's io.Writer
	wb, err := w.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write file data: %w", err)
	}

	if wb < len(data) {
		return ErrWriteIncomplete
	}

	log.Debug().
		Str("file", id).
		Msg(fmt.Sprintf("read %d bytes from memory", wb))

	return nil
}

// WriteObject writes a file to the local storage directory from the provided io.Reader
func (s *MemoryStore) WriteObject(id string, r io.Reader) error {
	// check the file does not already exist
	_, ok := s.objects.Load(id)
	if ok {
		return ErrFileExists
	}

	// read all the bytes from the buffer
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read bytes from request: %w", err)
	}

	// attempt to store it, fail if another writer beats us
	_, loaded := s.objects.LoadOrStore(id, data)
	if loaded {
		return ErrFileExists
	}

	log.Debug().
		Str("file", id).
		Msg(fmt.Sprintf("wrote %d bytes to memory", len(data)))

	return nil
}
