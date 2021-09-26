package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"net"
	"net/http"
	"testing"

	"github.com/purehyperbole/catly/protocol/catly"
	"github.com/purehyperbole/catly/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func testGRPCServer(t *testing.T, maxRequestSize int) (net.Listener, *storage.MemoryStore) {
	return testGRPCServerWithDetector(t, maxRequestSize, func(data []byte) string {
		return "image/jpeg"
	})
}

func testGRPCServerWithDetector(t *testing.T, maxRequestSize int, detector contentDetectorFunc) (net.Listener, *storage.MemoryStore) {
	listener, err := net.Listen("tcp", ":8000")
	require.NoError(t, err)

	m := storage.NewMemoryStore()

	s := grpc.NewServer(
		grpc.MaxSendMsgSize(maxRequestSize),
		grpc.MaxRecvMsgSize(maxRequestSize),
	)

	r := NewGRPCResource(
		"http://127.0.0.1:8080/",
		m,
	)

	r.contentDetector = detector

	catly.RegisterObjectServer(s, r)
	go s.Serve(listener)

	return listener, m
}

func testGRPCClient(t *testing.T) catly.ObjectClient {
	conn, err := grpc.Dial("127.0.0.1:8000", grpc.WithInsecure())
	require.NoError(t, err)

	return catly.NewObjectClient(conn)
}

func TestObjectUpload(t *testing.T) {
	s, m := testGRPCServer(t, 1<<20)
	c := testGRPCClient(t)
	defer s.Close()

	data := make([]byte, 1<<18)
	rand.Read(data)

	req := &catly.UploadObjectRequest{
		Name: "cat.jpg",
		Data: data,
	}

	resp, err := c.Upload(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, catly.ObjectStatus_ObjectOK, resp.Status)
	assert.Empty(t, resp.Error)
	assert.Equal(t, "http://127.0.0.1:8080/cat.jpg", resp.Url)

	var b bytes.Buffer

	err = m.ReadObject("cat.jpg", &b)
	require.NoError(t, err)
	assert.Equal(t, data, b.Bytes())
}

func TestObjectUploadNameConflict(t *testing.T) {
	s, m := testGRPCServer(t, 1<<20)
	c := testGRPCClient(t)
	defer s.Close()

	data1 := make([]byte, 1<<18)
	rand.Read(data1)

	data2 := make([]byte, 1<<18)
	rand.Read(data2)

	err := m.WriteObject("cat.jpg", bytes.NewReader(data1))
	require.Nil(t, err)

	req := &catly.UploadObjectRequest{
		Name: "cat.jpg",
		Data: data2,
	}

	resp, err := c.Upload(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, catly.ObjectStatus_ObjectERR, resp.Status)
	assert.Equal(t, storage.ErrFileExists.Error(), resp.Error)
	assert.Empty(t, resp.Url)

	var b bytes.Buffer

	err = m.ReadObject("cat.jpg", &b)
	require.NoError(t, err)
	assert.Equal(t, data1, b.Bytes())
}

func TestObjectUploadFileExtensionMismatch(t *testing.T) {
	s, _ := testGRPCServer(t, 1<<20)
	c := testGRPCClient(t)
	defer s.Close()

	data := make([]byte, 1<<18)
	rand.Read(data)

	req := &catly.UploadObjectRequest{
		Name: "cat.png",
		Data: data,
	}

	resp, err := c.Upload(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, catly.ObjectStatus_ObjectERR, resp.Status)
	assert.Equal(t, "uploaded image extension '.png' does not match it's content type of 'image/jpeg'", resp.Error)
	assert.Empty(t, resp.Url)
}

func TestObjectUploadUnsupportedContent(t *testing.T) {
	s, _ := testGRPCServerWithDetector(t, 1<<20, http.DetectContentType)
	c := testGRPCClient(t)
	defer s.Close()

	data := make([]byte, 1<<18)
	rand.Read(data)

	req := &catly.UploadObjectRequest{
		Name: "cat.jpg",
		Data: data,
	}

	resp, err := c.Upload(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, catly.ObjectStatus_ObjectERR, resp.Status)
	assert.Equal(t, "uploaded image content of 'application/octet-stream' is not supported", resp.Error)
	assert.Empty(t, resp.Url)
}

func TestObjectUploadNameTooLarge(t *testing.T) {
	s, _ := testGRPCServer(t, 1<<20)
	c := testGRPCClient(t)
	defer s.Close()

	data := make([]byte, 1<<18)
	rand.Read(data)

	name := make([]byte, 256)

	req := &catly.UploadObjectRequest{
		Name: string(name) + ".jpg",
		Data: data,
	}

	resp, err := c.Upload(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, catly.ObjectStatus_ObjectERR, resp.Status)
	assert.Equal(t, "image name should be between 1 and 256 characters", resp.Error)
	assert.Empty(t, resp.Url)
}

func TestObjectUploadNameInvalidCharacter(t *testing.T) {
	s, _ := testGRPCServer(t, 1<<20)
	c := testGRPCClient(t)
	defer s.Close()

	data := make([]byte, 1<<18)
	rand.Read(data)

	req := &catly.UploadObjectRequest{
		Name: "../../cat.jpg",
		Data: data,
	}

	resp, err := c.Upload(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, catly.ObjectStatus_ObjectERR, resp.Status)
	assert.Equal(t, "image name contains invalid characters", resp.Error)
	assert.Empty(t, resp.Url)
}

func TestObjectUploadNoData(t *testing.T) {
	s, _ := testGRPCServer(t, 1<<20)
	c := testGRPCClient(t)
	defer s.Close()

	data := make([]byte, 1<<18)
	rand.Read(data)

	req := &catly.UploadObjectRequest{
		Name: "cat.jpg",
	}

	resp, err := c.Upload(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, catly.ObjectStatus_ObjectERR, resp.Status)
	assert.Equal(t, "image upload contains no valid data", resp.Error)
	assert.Empty(t, resp.Url)
}

func TestObjectUploadFileTooLarge(t *testing.T) {
	s, _ := testGRPCServer(t, 1<<20)
	c := testGRPCClient(t)
	defer s.Close()

	data := make([]byte, 1<<22)
	rand.Read(data)

	req := &catly.UploadObjectRequest{
		Name: "../../cat.jpg",
		Data: data,
	}

	_, err := c.Upload(context.Background(), req)
	require.Error(t, err)
}
