package dataloader_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/graph-gophers/dataloader/v8"
)

func TestKeyOf(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		key := KeyOf[int](5)
		assert.Implements(t, (*Key[int])(nil), key)
		assert.Equal(t, "5", key.String())
		assert.Equal(t, 5, key.Raw())
	})

	t.Run("uint32", func(t *testing.T) {
		key := KeyOf[uint32](53)
		assert.Implements(t, (*Key[uint32])(nil), key)
		assert.Equal(t, "53", key.String())
		assert.Equal(t, uint32(53), key.Raw())
	})

	t.Run("[2]int", func(t *testing.T) {
		key := KeyOf([...]int{5, 3})
		assert.Implements(t, (*Key[[2]int])(nil), key)
		assert.Equal(t, "[5 3]", key.String())
		assert.Equal(t, [2]int{5, 3}, key.Raw())
	})

	t.Run("comparable struct", func(t *testing.T) {
		type foo struct {
			a int
			b string
		}

		raw := foo{a: 5, b: "bar"}
		key := KeyOf(raw)
		assert.Implements(t, (*Key[foo])(nil), key)
		assert.Equal(t, "{5 bar}", key.String())
		assert.Equal(t, raw, key.Raw())
	})
}

func TestKeysFrom(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		keys := KeysFrom[int](5, 6)
		assert.IsType(t, (Keys[int])(nil), keys)
		assert.Equal(t, []string{"5", "6"}, keys.Keys())
		assert.Equal(t, []int{5, 6}, keys.Raws())
	})

	t.Run("uint32", func(t *testing.T) {
		keys := KeysFrom[uint32](5, 3)
		assert.IsType(t, (Keys[uint32])(nil), keys)
		assert.Equal(t, []string{"5", "3"}, keys.Keys())
		assert.Equal(t, []uint32{5, 3}, keys.Raws())
	})

	t.Run("[2]int", func(t *testing.T) {
		keys := KeysFrom([...]int{5, 3}, [...]int{4, 9})
		assert.IsType(t, (Keys[[2]int])(nil), keys)
		assert.Equal(t, []string{"[5 3]", "[4 9]"}, keys.Keys())
		assert.Equal(t, [][2]int{{5, 3}, {4, 9}}, keys.Raws())
	})

	t.Run("comparable struct", func(t *testing.T) {
		type foo struct {
			a int
			b string
		}

		keys := KeysFrom(foo{a: 5, b: "bar"}, foo{a: 42, b: "foobar"})
		assert.IsType(t, (Keys[foo])(nil), keys)
		assert.Equal(t, []string{"{5 bar}", "{42 foobar}"}, keys.Keys())
		assert.Equal(t, []foo{{a: 5, b: "bar"}, {a: 42, b: "foobar"}}, keys.Raws())
	})
}

func TestStringerKey(t *testing.T) {
	raw := stringer{"foo", []int{42, 23}}
	key := StringerKey(raw)
	assert.Implements(t, (*Key[stringer])(nil), key)
	expectedID := raw.String()
	assert.Equalf(t, expectedID, key.String(), "String() value must match expected `%s`", expectedID)
	assert.Equal(t, raw, key.Raw())
}

func TestKeysFromStringers(t *testing.T) {
	raws := []stringer{
		{"foo", []int{42, 23}},
		{"bar", []int{4711, 1337, 1887}},
	}

	keys := KeysFromStringers(raws...)
	assert.IsType(t, (Keys[stringer])(nil), keys)
	assert.Equal(t, []string{
		"foo[42 23]",
		"bar[4711 1337 1887]",
	}, keys.Keys())
	assert.Equal(t, raws, keys.Raws())
}

// stringer represents struct, which is not comparable, but shall still be used as dataloader key
type stringer struct {
	name string
	// this field prevents this type to be used as comparable
	kpis []int
}

func (s stringer) String() string {
	return fmt.Sprintf("%s%v", s.name, s.kpis)
}
