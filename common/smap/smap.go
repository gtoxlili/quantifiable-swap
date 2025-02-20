package smap

import "sync"

type SyncMap[K comparable, V any] struct {
	m sync.Map
}

func (sm *SyncMap[K, V]) Load(key K) (value V, ok bool) {
	rawValue, ok := sm.m.Load(key)
	if ok {
		value = rawValue.(V)
	}
	return value, ok
}

func (sm *SyncMap[K, V]) Store(key K, value V) {
	sm.m.Store(key, value)
}

func (sm *SyncMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	rawActual, loaded := sm.m.LoadOrStore(key, value)
	actual = rawActual.(V)
	return actual, loaded
}

func (sm *SyncMap[K, V]) Delete(key K) {
	sm.m.Delete(key)
}

func (sm *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	sm.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}
