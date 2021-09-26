package storage

import "errors"

var (
	// ErrDirectoryPathIsFile is returned when the specified base directory is a file
	ErrDirectoryPathIsFile = errors.New("storage base directory path is a file")
	// ErrFileDoesNotExist is returned when a requested file cannot be found
	ErrFileDoesNotExist = errors.New("the file you requested does not exist")
	// ErrFileExist is returned when creating a file that already exists with the same filename
	ErrFileExists = errors.New("the file you have uploaded must have a unique name")
	// ErrWriteIncomplete is returned when write operation to a requesters io.Writer is incomplete
	ErrWriteIncomplete = errors.New("write incomplete")
)
