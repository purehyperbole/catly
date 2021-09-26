package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// FileStore an implementation of the object storage
// that writes to files in a directory
type FileStore struct {
	// the base storage directory where files will be stored
	baseDir string
}

// NewFileStore creates a new file store in the specified directory
func NewFileStore(baseDir string) (*FileStore, error) {
	s, err := os.Stat(baseDir)
	if err != nil {
		return nil, fmt.Errorf("storage base directory does not exist: %w", err)
	}

	if !s.IsDir() {
		return nil, ErrDirectoryPathIsFile
	}

	// TODO : check if directory is writable

	return &FileStore{
		baseDir: baseDir,
	}, nil
}

// ReadObject reads a file from the local storage directory to the provided io.Writer
func (s *FileStore) ReadObject(id string, w io.Writer) error {
	p := filepath.Join(s.baseDir, id)

	// Here we open the file and return it as an io.Reader so it's contents can
	// be streamed to the requester
	fd, err := os.OpenFile(p, os.O_RDONLY, 0644)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrFileDoesNotExist
		}

		return fmt.Errorf("failed to read requested file: %w", err)
	}

	defer fd.Close()

	rb, err := io.Copy(w, fd)
	if err != nil {
		return err
	}

	log.Debug().
		Str("file", id).
		Str("directory", s.baseDir).
		Msg(fmt.Sprintf("read %d bytes from disk", rb))

	return nil
}

// WriteObject writes a file to the local storage directory from the provided io.Reader
func (s *FileStore) WriteObject(id string, r io.Reader) error {
	p := filepath.Join(s.baseDir, id)

	// we open the file with O_CREATE and O_EXCL, which will prevent race conditions
	// when someone uploads a file with the same name as us at the same time
	fd, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return ErrFileExists
		}

		return err
	}

	defer fd.Close()

	// copy data from the request's body to the file descriptor
	wb, err := io.Copy(fd, r)
	if err != nil {
		return fmt.Errorf("file upload failed: %w", err)
	}

	log.Debug().
		Str("file", id).
		Str("directory", s.baseDir).
		Msg(fmt.Sprintf("wrote %d bytes to disk", wb))

	return nil
}
