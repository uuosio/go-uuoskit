package uuoskit

import (
	"github.com/iancoleman/orderedmap"
)

func DeepGet(m *orderedmap.OrderedMap, keys ...interface{}) (interface{}, bool) {
	value, ok := m.Get(keys[0].(string))
	if !ok {
		return nil, false
	}

	for _, key := range keys[1:] {
		if _key, ok := key.(string); ok {
			if _value, ok := value.(orderedmap.OrderedMap); ok {
				value, ok = _value.Get(_key)
				if !ok {
					return nil, false
				}
			} else {
				return nil, false
			}
		} else if _key, ok := key.(int); ok {
			if _value, ok := value.([]interface{}); ok {
				if _key >= len(_value) {
					return nil, false
				}
				value = _value[_key]
			} else {
				return nil, false
			}
		}
	}
	return value, true
}
