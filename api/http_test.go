package api

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/purehyperbole/catly/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testHTTPResource(t *testing.T) (*HTTPResource, *storage.MemoryStore) {
	m := storage.NewMemoryStore()
	r := NewHTTPResource(m)
	return r, m
}

func TestHTTPGetObject(t *testing.T) {
	r, m := testHTTPResource(t)

	data := make([]byte, 1024)
	rand.Read(data)

	err := m.WriteObject("cat.jpg", bytes.NewReader(data))
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "/cat.jpg", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()

	r.GetObject(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "image/jpeg", rec.Header().Get("Content-Type"))
	assert.Equal(t, data, rec.Body.Bytes())
}

func TestHTTPGetObjectBadMethod(t *testing.T) {
	r, m := testHTTPResource(t)

	data := make([]byte, 1024)
	rand.Read(data)

	err := m.WriteObject("cat.jpg", bytes.NewReader(data))
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/cat.jpg", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()

	r.GetObject(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestHTTPGetObjectDoesntExist(t *testing.T) {
	r, _ := testHTTPResource(t)

	req, err := http.NewRequest(http.MethodGet, "/cat.jpg", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()

	r.GetObject(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHTTPGetObjectNameTooLarge(t *testing.T) {
	r, _ := testHTTPResource(t)

	fileID := hex.EncodeToString(make([]byte, 512))

	req, err := http.NewRequest(http.MethodGet, "/"+fileID, nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()

	r.GetObject(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHTTPGetObjectNameInvalidCharacter(t *testing.T) {
	r, _ := testHTTPResource(t)

	req, err := http.NewRequest(http.MethodGet, "/../../cat.jpg", nil)
	require.NoError(t, err)

	rec := httptest.NewRecorder()

	r.GetObject(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
