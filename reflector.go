package exql

import (
	"fmt"
	"reflect"

	"github.com/loilo-inc/exql/v3/util"
)

var errModelNil = fmt.Errorf("model is nil")

type reflector struct {
	metadata util.SyncMap[string, *modelSchema]
}

type Reflector interface {
	GetSchema(modelPtr any) (*modelSchema, error)
	GetSchemaFromValue(destValue *reflect.Value) (*modelSchema, error)
}

var _ Reflector = (*reflector)(nil)

func (r *reflector) GetSchema(modelPtr any) (*modelSchema, error) {
	if modelPtr == nil {
		return nil, errModelNil
	}
	value, err := resolveDestination(modelPtr)
	if err != nil {
		return nil, err
	}
	return r.GetSchemaFromValue(value)
}

func (r *reflector) GetSchemaFromValue(destValue *reflect.Value) (*modelSchema, error) {
	destType, err := resolveDestType(destValue)
	if err != nil {
		return nil, err
	}
	key := typeKey(destType)
	if v, ok := r.metadata.Load(key); ok {
		return v, nil
	}
	f, err := aggregateFields(destType)
	if err != nil {
		return nil, err
	}
	r.metadata.Store(key, f)
	return f, nil
}

func typeKey(t reflect.Type) string {
	return t.PkgPath() + "." + t.Name()
}

type noCacheReflector struct{}

var _ Reflector = (*noCacheReflector)(nil)

func defaultReflector() Reflector {
	return &noCacheReflector{}
}

func (r *noCacheReflector) GetSchema(modelPtr any) (*modelSchema, error) {
	if modelPtr == nil {
		return nil, errModelNil
	}
	value, err := resolveDestination(modelPtr)
	if err != nil {
		return nil, err
	}
	return r.GetSchemaFromValue(value)
}

func (r *noCacheReflector) GetSchemaFromValue(destValue *reflect.Value) (*modelSchema, error) {
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
		return nil, errModelNil
	}
	return destType, nil
}
