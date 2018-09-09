package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/datskos/ratelimit/pkg/proto"
	"github.com/datskos/ratelimit/pkg/storage"
	"github.com/stretchr/testify/assert"
)

// test adjustments (refills) _before_ reducing tokens
func TestRefill(t *testing.T) {
	var tbl = []struct {
		curr      uint32
		max       uint32
		timeDiff  uint32
		refillSec uint32
		refillAmt uint32
		exp       uint32
	}{
		{5, 5, 10, 10, 5, 5}, // already full
		{0, 5, 10, 10, 5, 5}, // full refill
		{0, 5, 30, 10, 5, 5}, // full refill

		{0, 40, 10, 10, 5, 5},  // partial refill
		{0, 40, 19, 10, 5, 5},  // partial refill
		{0, 40, 20, 10, 5, 10}, // partial refill
		{0, 40, 80, 10, 5, 40}, // partial refill
		{0, 40, 90, 10, 5, 40}, // partial refill

		{0, 5, 5, 10, 5, 0}, // zero refill
		{0, 5, 2, 10, 5, 0}, // zero refill
		{0, 5, 1, 10, 5, 0}, // zero refill
	}

	for i, tt := range tbl {
		now := time.Now().UTC()
		tLast := now.Add(-1 * time.Duration(tt.timeDiff) * time.Second)
		value := &storage.Value{
			Remaining:      tt.curr,
			LastRefilledAt: tLast,
		}
		params := &params{
			maxAmount:      tt.max,
			refillAmount:   tt.refillAmt,
			refillDuration: time.Duration(tt.refillSec) * time.Second,
			now:            now,
		}
		params.adjustTokens(value)
		assert.Equal(t, tt.exp, value.Remaining,
			fmt.Sprintf("remaining and expected tokens differ. idx=%d", i))
	}

}

// test token reduction exclusively of token refills
func TestReduce(t *testing.T) {
	var tbl = []struct {
		curr         uint32
		expRemaining uint32
		expStatus    proto.ReduceResponse_Status
	}{
		{10, 9, proto.ReduceResponse_OK},
		{1, 0, proto.ReduceResponse_OK},
		{0, 0, proto.ReduceResponse_NG},
	}

	for i, tt := range tbl {
		value := &storage.Value{Remaining: tt.curr}
		params := &params{}
		status := params.reduce(value)
		assert.Equal(t, tt.expStatus, status, fmt.Sprintf("exp status wrong. id=%d", i))
		assert.Equal(t, tt.expRemaining, value.Remaining, fmt.Sprintf("exp count wrong. id=%d", i))
	}

}
