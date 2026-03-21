package exql

import (
	"fmt"
	"reflect"

	"github.com/loilo-inc/exql/v3/util"
	"golang.org/x/xerrors"
)

var errModelTypeNil = fmt.Errorf("model type is nil")

type reflector struct {
	fields util.SyncMap[string, *util.SyncMap[string, int]]
}

type Reflector interface {
	GetFields(destValue *reflect.Value) (*util.SyncMap[string, int], error)
}

var _ Reflector = (*reflector)(nil)

func (r *reflector) GetFields(destValue *reflect.Value) (*util.SyncMap[string, int], error) {
	destType, err := resolveDestType(destValue)
	if err != nil {
		return nil, err
	}
	key := typeKey(destType)
	if v, ok := r.fields.Load(key); ok {
		return v, nil
	}
	f, err := aggregateFields(destType)
	if err != nil {
		return nil, err
	}
	r.fields.Store(key, f)
	return f, nil
}

func typeKey(t reflect.Type) string {
	return t.PkgPath() + "." + t.Name()
}

type noCacheReflector struct{}

var _ Reflector = (*noCacheReflector)(nil)
var defaultReflector = &noCacheReflector{}

func (r *noCacheReflector) GetFields(destValue *reflect.Value) (*util.SyncMap[string, int], error) {
	destType, err := resolveDestType(destValue)
	if err != nil {
		return nil, err
	}
	return aggregateFields(destType)
}

func resolveDestType(destValue *reflect.Value) (reflect.Type, error) {
	// *Model || **Model
	destType := destValue.Type()
	if destValue.Kind() == reflect.Pointer {
		destType = destType.Elem()
	}
	if destType == nil {
		return nil, errModelTypeNil
	}
	return destType, nil
}

func aggregateFields(t reflect.Type) (*util.SyncMap[string, int], error) {
	fields := &util.SyncMap[string, int]{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("exql")
		if tag != "" {
			if f.Type.Kind() == reflect.Pointer {
				return nil, xerrors.Errorf("struct field must not be a pointer: %s %s", f.Type.Name(), f.Type.Kind())
			}
			tags, err := ParseTags(tag)
			if err != nil {
				return nil, err
			}
			col := tags["column"]
			fields.Store(col, i)
		}
	}
	return fields, nil
}
