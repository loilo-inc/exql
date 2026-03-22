package exql

import (
	"fmt"
	"reflect"
)

var errModelNil = fmt.Errorf("model is nil")

type reflector struct {
	noCache bool
	schemas syncMap[string, *modelSchema]
}

// Reflector is an interface to manage model metadata used for query generation and mapping.
type Reflector interface {
	// GetSchema returns the model schema for the given model pointer.
	GetSchema(modelPtr any) (*modelSchema, error)
	// GetSchemaFromValue returns the model schema for the given reflect.Value of the destination struct.
	GetSchemaFromValue(destValue *reflect.Value) (*modelSchema, error)
	// ClearSchemaCache clears the cached model schemas.
	ClearSchemaCache()
}

var _ Reflector = (*reflector)(nil)

func (r *reflector) GetSchema(modelPtr any) (*modelSchema, error) {
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
	if !r.noCache {
		if v, ok := r.schemas.Load(key); ok {
			return v, nil
		}
	}
	f, err := aggregateFields(destType)
	if err != nil {
		return nil, err
	}
	if !r.noCache {
		r.schemas.Store(key, f)
	}
	return f, nil
}

func (r *reflector) ClearSchemaCache() {
	r.schemas.m.Clear()
}

func typeKey(t reflect.Type) string {
	return t.PkgPath() + "." + t.Name()
}

var noCacheReflector = &reflector{noCache: true}

// resolveDestination validates that the input is a pointer to a struct and returns the reflect.Value of the struct.
func resolveDestination(pointerOfStruct any) (*reflect.Value, error) {
	if pointerOfStruct == nil {
		return nil, errMapDestination
	}
	destValue := reflect.ValueOf(pointerOfStruct)
	// any -> (*Model)
	destType := destValue.Type()
	if destType.Kind() != reflect.Pointer {
		return nil, errMapDestination
	}
	// (*Model) -> Model
	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Struct {
		return nil, errMapDestination
	}
	return &destValue, nil
}

var errMapRowSerialDestination = fmt.Errorf("destination must be either *(struct) or *((*struct)(nil))")

// resolveNullableDestination validates that the input is a pointer to a struct or a pointer to a pointer to a struct and returns the reflect.Value of the struct.
func resolveNullableDestination(dest any) (*reflect.Value, error) {
	if dest == nil {
		return nil, errMapRowSerialDestination
	}
	// any -> (*Model) || (**Model)
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Pointer {
		return nil, errMapRowSerialDestination
	}
	destValue = destValue.Elem()
	switch destValue.Kind() {
	case reflect.Struct:
		// *Model -> Model
		return &destValue, nil
	case reflect.Pointer:
		// **Model -> *Model (nil only)
		if destValue.Type().Elem().Kind() != reflect.Struct {
			return nil, errMapRowSerialDestination
		}
		if !destValue.IsNil() {
			return nil, errMapRowSerialDestination
		}
		return &destValue, nil
	}
	return nil, errMapRowSerialDestination
}

var errMapManyDestination = fmt.Errorf("destination must be a pointer of slice of struct")

// resolveDestinationMany validates that the input is a pointer to a slice of pointers to struct and returns the reflect.Type of the struct and the reflect.Value of the destination slice.
func resolveDestinationMany(ptrOfSliceOfModelPtr any) (reflect.Type, *reflect.Value, error) {
	if ptrOfSliceOfModelPtr == nil {
		return nil, nil, errMapManyDestination
	}
	destValue := reflect.ValueOf(ptrOfSliceOfModelPtr)
	// any -> *[]*Model
	destType := destValue.Type()
	if destType.Kind() != reflect.Pointer {
		return nil, nil, errMapManyDestination
	}
	// *[]*Model -> []*Model
	destType = destType.Elem()
	if destType.Kind() != reflect.Slice {
		return nil, nil, errMapManyDestination
	}
	// []*Model -> *Model
	sliceType := destType.Elem()
	if sliceType.Kind() != reflect.Pointer {
		return nil, nil, errMapManyDestination
	}
	// *Model -> Model
	sliceType = sliceType.Elem()
	if sliceType.Kind() != reflect.Struct {
		return nil, nil, errMapManyDestination
	}
	return sliceType, &destValue, nil
}

// resolveDestType returns the reflect.Type of the destination struct, handling both *Model and **Model cases.
func resolveDestType(destValue *reflect.Value) (reflect.Type, error) {
	if !destValue.IsValid() {
		return nil, errModelNil
	}
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
