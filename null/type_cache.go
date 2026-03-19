package null

import (
	"reflect"
	"sync"
)

var textTypeCache sync.Map

type textTypeInfo struct {
	isString bool
}

func textTypeInfoFor[T any]() textTypeInfo {
	t := reflect.TypeFor[T]()
	if info, ok := textTypeCache.Load(t); ok {
		return info.(textTypeInfo)
	}
	info := textTypeInfo{isString: t.Kind() == reflect.String}
	actual, _ := textTypeCache.LoadOrStore(t, info)
	return actual.(textTypeInfo)
}
