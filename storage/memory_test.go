package storage

import (
	"bytes"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMemoryStore(t *testing.T) *MemoryStore {
	return NewMemoryStore()
}

func TestMemoryStorageReadFile(t *testing.T) {
	fs := newTestMemoryStore(t)

	fs.objects.Store("cat.jpg", []byte("meow"))

	var b bytes.Buffer

	err := fs.ReadObject("cat.jpg", &b)
	require.NoError(t, err)
	assert.Equal(t, []byte("meow"), b.Bytes())
}

func TestMemoryStorageReadFileNotExist(t *testing.T) {
	// test requesting a file that does not exist
	fs := newTestMemoryStore(t)

	var b bytes.Buffer

	err := fs.ReadObject("invisible-cat.jpg", &b)
	require.Equal(t, ErrFileDoesNotExist, err)
	assert.Equal(t, 0, b.Len())
}

func TestMemoryStorageWriteFile(t *testing.T) {
	fs := newTestMemoryStore(t)

	r := bytes.NewReader([]byte("meow"))

	err := fs.WriteObject("cat.jpg", r)
	require.NoError(t, err)

	value, ok := fs.objects.Load("cat.jpg")
	require.True(t, ok)

	data, ok := value.([]byte)
	require.True(t, ok)
	assert.Equal(t, []byte("meow"), data)
}

func TestMemoryStorageWriteConcurrentConflict(t *testing.T) {
	fs := newTestMemoryStore(t)

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
	value, ok := fs.objects.Load("cat.jpg")
	require.True(t, ok)

	data, ok := value.([]byte)
	require.True(t, ok)
	assert.Equal(t, []byte("meow"), data)

	// check the other 99 requests failed
	assert.Equal(t, int64(99), errorCount)
}
