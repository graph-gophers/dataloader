package dataloader

// Key is the interface that all keys need to implement
type Key interface {
	// Key returns a guaranteed unique string that can be used to identify an object
	Key() string
	// Raw returns the raw, underlaying value of the key
	Raw() interface{}
}

// KeyList is a slice of keys with some convience methods
type Keys []Key

// Keys returns the list of strings. One for each "Key" in the list
func (l Keys) Keys() []string {
	list := make([]string, len(l))
	for i := range l {
		list[i] = l[i].Key()
	}
	return list
}

// StringKey implements the Key interface for a string
type StringKey string

// Key is an identity method. Used to implement Key interface
func (k StringKey) Key() string { return string(k) }

// String is an identity method. Used to implement Key Raw
func (k StringKey) Raw() interface{} { return k }

// NewKeysFromStrings converts a `[]strings` to a `Keys` ([]Key)
func NewKeysFromStrings(strings []string) Keys {
	list := make(Keys, len(strings))
	for i := range strings {
		list[i] = StringKey(strings[i])
	}
	return list
}
