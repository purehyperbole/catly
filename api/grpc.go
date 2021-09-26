package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/purehyperbole/catly/protocol/catly"
)

type contentDetectorFunc func(data []byte) string

// WritableStorage specifies the interface that storage
// backends will need to implement for the gRPC api
type WritableStorage interface {
	WriteObject(id string, r io.Reader) error
}

// GRPCResource an implementation of the gRPC object service
type GRPCResource struct {
	address         string
	storage         WritableStorage
	contentDetector contentDetectorFunc
}

// NewGRPCResource creates a new grpc implementation of the object service
func NewGRPCResource(address string, ws WritableStorage) *GRPCResource {
	return &GRPCResource{
		address:         address,
		storage:         ws,
		contentDetector: http.DetectContentType,
	}
}

// Upload handles upload requests for images
func (rs *GRPCResource) Upload(ctx context.Context, req *catly.UploadObjectRequest) (*catly.UploadObjectResponse, error) {
	// check the name of the file is present and not too large
	if len(req.Name) > 256 || len(req.Name) < 1 {
		return &catly.UploadObjectResponse{
			Status: catly.ObjectStatus_ObjectERR,
			Error:  "image name should be between 1 and 256 characters",
		}, nil
	}

	// check there are no slashes to prevent someone from trying to escape
	// to other parts of the filesystem (if file storage is used)
	if strings.ContainsRune(req.Name, '/') {
		return &catly.UploadObjectResponse{
			Status: catly.ObjectStatus_ObjectERR,
			Error:  "image name contains invalid characters",
		}, nil
	}

	// check that data has been provided
	if len(req.Data) < 1 {
		return &catly.UploadObjectResponse{
			Status: catly.ObjectStatus_ObjectERR,
			Error:  "image upload contains no valid data",
		}, nil
	}

	// get a best effort guess at the data's contents
	mt := rs.contentDetector(req.Data)

	if mt != "image/jpeg" && mt != "image/png" && mt != "image/gif" {
		return &catly.UploadObjectResponse{
			Status: catly.ObjectStatus_ObjectERR,
			Error:  fmt.Sprintf("uploaded image content of '%s' is not supported", mt),
		}, nil
	}

	// check that the detected mime type matches the file extension
	// provided by the user
	ext := filepath.Ext(req.Name)

	if mt != mime.TypeByExtension(ext) {
		return &catly.UploadObjectResponse{
			Status: catly.ObjectStatus_ObjectERR,
			Error:  fmt.Sprintf("uploaded image extension '%s' does not match it's content type of '%s'", ext, mt),
		}, nil
	}

	// write the object to the underlying storage implementation
	err := rs.storage.WriteObject(req.Name, bytes.NewReader(req.Data))
	if err != nil {
		return &catly.UploadObjectResponse{
			Status: catly.ObjectStatus_ObjectERR,
			Error:  err.Error(),
		}, nil
	}

	// generate the URL and return it to the uploader
	return &catly.UploadObjectResponse{
		Status: catly.ObjectStatus_ObjectOK,
		Url:    fmt.Sprintf("%s%s", rs.address, req.Name),
	}, nil
}
