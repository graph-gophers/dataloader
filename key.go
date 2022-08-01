package dataloader

import "fmt"

// Key is the interface that all keys need to implement
type Key[K any] interface {
	fmt.Stringer
	// Raw returns the underlying value of the key
	Raw() K
}

// Keys wraps a slice of Key types to provide some convenience methods.
type Keys[K any] []Key[K]

// Keys returns the list of strings. One for each "Key" in the list
func (l Keys[K]) Keys() []string {
	list := make([]string, len(l))
	for i := range l {
		list[i] = l[i].String()
	}
	return list
}

// Raws returns the list of raw values in the key list
func (l Keys[K]) Raws() []K {
	list := make([]K, len(l))
	for i := range l {
		list[i] = l[i].Raw()
	}
	return list
}

// KeyOf wraps the given comparable type as Key
func KeyOf[K comparable](item K) Key[K] {
	return comparableKey[K]{item}
}

// KeysFrom wraps a variadic list of comparable types as Keys
func KeysFrom[K comparable](items ...K) Keys[K] {
	list := make(Keys[K], len(items))
	for i := range items {
		list[i] = comparableKey[K]{items[i]}
	}

	return list
}

// StringerKey wraps the given fmt.Stringer implementation as Key, so it can be used in the dataloader
// The Key ist strictly typed to the implementing type, and cannot be mixed with other Keys,
// which themselves implement the fmt.Stringer interface
func StringerKey[K fmt.Stringer](item K) Key[K] {
	return stringerKey[K]{item}
}

// KeysFromStringers wraps the given variadic list of fmt.Stringer implementations as Keys
// The Keys are strictly typed to the implementing type of the first element
// The normal rules of type inference on type parameters apply
func KeysFromStringers[K fmt.Stringer](items ...K) Keys[K] {
	list := make(Keys[K], len(items))
	for i := range items {
		list[i] = stringerKey[K]{items[i]}
	}

	return list
}

// comparableKey implements the Key interface for any comparable type
type comparableKey[K comparable] struct {
	cmp K
}

func (k comparableKey[K]) String() string {
	return fmt.Sprintf("%v", k.cmp)
}

func (k comparableKey[K]) Raw() K {
	return k.cmp
}

var _ Key[Key[string]] = (*stringerKey[Key[string]])(nil)

type stringerKey[K fmt.Stringer] struct {
	raw K
}

func (k stringerKey[K]) String() string {
	return k.raw.String()
}

func (k stringerKey[K]) Raw() K {
	return k.raw
}
