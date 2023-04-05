package dataloader

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	type UserKey struct {
		ID int
	}

	keys := []UserKey{{10}, {15}, {30}}
	kkeys := make([]Key[UserKey], len(keys))

	for i := range keys {
		kkeys[i] = ContextKey(context.Background(), keys[i])
	}

	raws := Keys[UserKey](kkeys).Raw()

	assert.EqualValues(t, raws, keys)
}
