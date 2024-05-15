package xload

import (
	"sync"
)

type collisionSyncMap sync.Map

func (cm *collisionSyncMap) add(key string) {
	m := (*sync.Map)(cm)
	v, loaded := m.LoadOrStore(key, 1)

	if loaded {
		m.Store(key, v.(int)+1)
	}
}

func (cm *collisionSyncMap) err() error {
	var collidedKeys []string

	m := (*sync.Map)(cm)
	m.Range(func(key, v any) bool {
		if key == "" {
			return true
		}

		if count, _ := v.(int); count > 1 {
			collidedKeys = append(collidedKeys, key.(string))
		}

		return true
	})

	return keysToErr(collidedKeys)
}

type collisionMap map[string]int

func (cm collisionMap) add(key string) { cm[key]++ }

func (cm collisionMap) err() error {
	var collidedKeys []string

	for key, count := range cm {
		if key == "" {
			continue
		}

		if count > 1 {
			collidedKeys = append(collidedKeys, key)
		}
	}

	return keysToErr(collidedKeys)
}

func keysToErr(collidedKeys []string) error {
	if len(collidedKeys) == 0 {
		return nil
	}

	return &ErrCollision{keys: collidedKeys}
}
