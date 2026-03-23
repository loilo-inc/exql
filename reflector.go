package exql

import (
	"fmt"
	"reflect"
	"sync"
)

var errModelNil = fmt.Errorf("model is nil")

type reflector struct {
	noCache bool
	schemas map[string]*modelSchema
	mux     sync.Mutex
}

func newReflector() Reflector {
	return &reflector{
		schemas: make(map[string]*modelSchema),
	}
}

// Reflector is an interface to manage model metadata used for query generation and mapping.
type Reflector interface {
	getSchema(destType reflect.Type, forUpdate bool) (*modelSchema, error)
	getModelSchema(dest any, forUpdate bool) (*modelSchema, error)
	clearSchemaCache()
}

var _ Reflector = (*reflector)(nil)

func (r *reflector) getModelSchema(dest any, forUpdate bool) (*modelSchema, error) {
	t, err := resolveModelType(dest)
	if err != nil {
		return nil, err
	}
	return r.getSchema(t, forUpdate)
}

func (r *reflector) getSchema(destType reflect.Type, forUpdate bool) (*modelSchema, error) {
	key := typeKey(destType)
	if !r.noCache {
		if v, ok := r.schemas[key]; ok {
			return v, nil
		}
	}
	f, err := aggregateFields(destType, forUpdate)
	if err != nil {
		return nil, err
	}
	if !r.noCache {
		r.mux.Lock()
		r.schemas[key] = f
		r.mux.Unlock()
	}
	return f, nil
}

func (r *reflector) clearSchemaCache() {
	r.mux.Lock()
	r.schemas = make(map[string]*modelSchema)
	r.mux.Unlock()
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

type nullableDest struct {
	elemType reflect.Type
	value    *reflect.Value
}

// resolveNullableDestination validates that the input is a pointer to a struct or a pointer to a pointer to a struct and returns the reflect.Value of the struct.
func resolveNullableDestination(dest any) (*nullableDest, error) {
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
		return &nullableDest{
			elemType: destValue.Type(),
			value:    &destValue,
		}, nil
	case reflect.Pointer:
		// **Model -> *Model (nil only)
		elemType := destValue.Type().Elem()
		if elemType.Kind() != reflect.Struct {
			return nil, errMapRowSerialDestination
		}
		if !destValue.IsNil() {
			return nil, errMapRowSerialDestination
		}
		return &nullableDest{
			elemType: elemType,
			value:    &destValue,
		}, nil
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

func resolveModelType(destValue any) (reflect.Type, error) {
	destType := reflect.TypeOf(destValue)
	if destType == nil {
		return nil, errModelNil
	}
	if destType.Kind() != reflect.Pointer {
		return nil, errMapDestination
	}
	return destType.Elem(), nil
}
