package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	value := &Value{
		Remaining:      100,
		LastRefilledAt: time.Now().UTC(),
		LastReducedAt:  time.Now().UTC(),
	}
	data, err := encode(value)
	require.NoError(t, err)

	decodedValue, err := decode(data)
	require.NoError(t, err)

	assert.Equal(t, value, decodedValue, "encoded/decoded values differ")
}
