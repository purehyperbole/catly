package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testStorageDirFile = "/tmp/im-a-file"
)

func newTestFileStore(t *testing.T) *FileStore {
	d, err := os.MkdirTemp("/tmp", "storage-*")
	require.NoError(t, err)

	fs, err := NewFileStore(d)
	require.NoError(t, err)

	return fs
}

func TestFileStorageSetup(t *testing.T) {
	// specify a folder that does not exist
	fs, err := NewFileStore("/tmp/does-not-exist")
	require.Error(t, err)
	assert.Nil(t, fs)

	// specify a file as the storage dir
	err = os.WriteFile(testStorageDirFile, []byte(`i'm a file!`), 0644)
	require.Nil(t, err)

	defer os.Remove(testStorageDirFile)

	fs, err = NewFileStore(testStorageDirFile)
	require.Error(t, err)
	assert.Equal(t, ErrDirectoryPathIsFile, err)
	assert.Nil(t, fs)

	// specify a correct storage dir
	d, err := os.MkdirTemp("/tmp", "storage-*")
	require.NoError(t, err)

	defer os.Remove(d)

	fs, err = NewFileStore(d)
	require.NoError(t, err)
	assert.NotNil(t, fs)
}

func TestFileStorageReadFile(t *testing.T) {
	fs := newTestFileStore(t)
	defer os.RemoveAll(fs.baseDir)

	err := os.WriteFile(filepath.Join(fs.baseDir, "cat.jpg"), []byte("meow"), 0644)
	require.NoError(t, err)

	var b bytes.Buffer

	err = fs.ReadObject("cat.jpg", &b)
	require.NoError(t, err)
	assert.Equal(t, []byte("meow"), b.Bytes())
}

func TestFileStorageReadFileNotExist(t *testing.T) {
	// test requesting a file that does not exist
	fs := newTestFileStore(t)
	defer os.RemoveAll(fs.baseDir)

	var b bytes.Buffer

	err := fs.ReadObject("invisible-cat.jpg", &b)
	require.Equal(t, ErrFileDoesNotExist, err)
	assert.Equal(t, 0, b.Len())
}

func TestFileStorageWriteFile(t *testing.T) {
	fs := newTestFileStore(t)
	defer os.RemoveAll(fs.baseDir)

	r := bytes.NewReader([]byte("meow"))

	err := fs.WriteObject("cat.jpg", r)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(fs.baseDir, "cat.jpg"))
	require.NoError(t, err)
	assert.Equal(t, []byte("meow"), data)
}

func TestFileStorageWriteConcurrentConflict(t *testing.T) {
	fs := newTestFileStore(t)
	defer os.RemoveAll(fs.baseDir)

	var wg sync.WaitGroup
	var errorCount int64

	wg.Add(100)

	// create 100 concurrent write requests
	// we expect that 99 of them should fail
	// as the file will already be created
	for i := 0; i < 100; i++ {
		go func() {
			r := bytes.NewReader([]byte("meow"))

			err := fs.WriteObject("cat.jpg", r)
			if err != nil {
				atomic.AddInt64(&errorCount, 1)
			}

			wg.Done()
		}()
	}

	wg.Wait()

	// check the file was uploaded
	data, err := os.ReadFile(filepath.Join(fs.baseDir, "cat.jpg"))
	require.NoError(t, err)
	assert.Equal(t, []byte("meow"), data)

	// check the other 99 requests failed
	assert.Equal(t, int64(99), errorCount)
}
