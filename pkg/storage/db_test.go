package storage

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/datskos/ratelimit/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runStorageTest(t *testing.T, test func(t *testing.T, storage Storage)) {
	dir, err := ioutil.TempDir("", "ratelimit-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	config := config.AppConfig{Port: 8080, DatabaseDir: dir}
	storage, err := NewStorage(config)
	require.NoError(t, err)
	defer storage.Close()

	test(t, storage)
}

func TestStorage(t *testing.T) {
	runStorageTest(t, func(t *testing.T, storage Storage) {
		key := "key:blah:blue"
		tx := storage.Tx()
		dne, err := tx.Get(key)
		assert.Nil(t, dne)
		assert.Nil(t, err)

		value := &Value{
			Remaining:      100,
			LastRefilledAt: time.Now().UTC(),
			LastReducedAt:  time.Now().UTC(),
		}
		err = tx.Set(key, value)
		require.NoError(t, err)
		err = tx.Commit()
		require.NoError(t, err)

		retrievedVal, err := storage.Get(key)
		assert.Equal(t, value, retrievedVal, "set/get values differ")
		assert.Nil(t, err)
	})
}
