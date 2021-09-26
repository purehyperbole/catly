package api

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/purehyperbole/catly/storage"
	"github.com/rs/zerolog/log"
)

// ReadableStorage specifies the interface that storage
// backends will need to implement for the http/web api
type ReadableStorage interface {
	ReadObject(id string, w io.Writer) error
}

// HTTPResource def
type HTTPResource struct {
	storage ReadableStorage
}

// NewHTTPResource creates a new server for http calls
func NewHTTPResource(s ReadableStorage) *HTTPResource {
	return &HTTPResource{
		storage: s,
	}
}

// GetObject handles GET requests for an object
func (rs *HTTPResource) GetObject(w http.ResponseWriter, r *http.Request) {
	id := uuid.New().String()

	log.Info().
		Str("id", id).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg("object requested")

	if r.Method != http.MethodGet {
		log.Warn().
			Str("id", id).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("bad request method")

		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}

	fileID := strings.TrimPrefix(r.URL.Path, "/")

	// fail if someone is trying to potentially access different paths
	// on the filesystem, or the image name is too large.
	// we should probably do more to sanitise the fileID here, but
	// it should be good for now
	if strings.ContainsRune(fileID, '/') || len(fileID) > 256 {
		log.Warn().
			Str("id", id).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("bad fileID requested")

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request: image URL is invalid"))
		return
	}

	// set the content type, get the object and write it to the response
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(fileID)))

	err := rs.storage.ReadObject(fileID, w)
	if err != nil {
		if errors.Is(err, storage.ErrFileDoesNotExist) {
			log.Warn().
				Str("id", id).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("requested file not found")

			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("image not found"))
			return
		}

		log.Warn().
			Str("id", id).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("error", err.Error()).
			Msg("could not serve file")

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
		return
	}

}
