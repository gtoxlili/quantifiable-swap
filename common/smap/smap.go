package smap

import "sync"

// SyncMap is a generic wrapper for sync.Map with type-safety for keys (K) and values (V).
type SyncMap[K comparable, V any] struct {
	m sync.Map
}

// Load returns the value stored in the map for a key, or false if no
// value is present.
func (sm *SyncMap[K, V]) Load(key K) (value V, ok bool) {
	rawValue, ok := sm.m.Load(key)
	if ok {
		value = rawValue.(V)
	}
	return value, ok
}

// Store sets the value for a key.
func (sm *SyncMap[K, V]) Store(key K, value V) {
	sm.m.Store(key, value)
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value. The loaded result
// is true if the value was loaded, false if it was stored.
func (sm *SyncMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	rawActual, loaded := sm.m.LoadOrStore(key, value)
	actual = rawActual.(V)
	return actual, loaded
}

// Delete deletes the value for a key.
func (sm *SyncMap[K, V]) Delete(key K) {
	sm.m.Delete(key)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (sm *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	sm.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}
