package exql

import (
	"reflect"
	"sync"
	"testing"

	"github.com/loilo-inc/exql/v3/model"
	"github.com/stretchr/testify/assert"
)

func Test_resolveDestination(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		_, err := resolveDestination(nil)
		assert.ErrorIs(t, errMapDestination, err)
	})
	t.Run("non-pointer", func(t *testing.T) {
		_, err := resolveDestination(1)
		assert.ErrorIs(t, errMapDestination, err)
	})
	t.Run("pointer of non-struct", func(t *testing.T) {
		var i int
		_, err := resolveDestination(&i)
		assert.ErrorIs(t, errMapDestination, err)
	})
	t.Run("*struct", func(t *testing.T) {
		u := &model.Users{Id: 1, Name: "alice"}
		v, err := resolveDestination(u)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), v.FieldByName("Id").Int())
		assert.Equal(t, "alice", v.FieldByName("Name").String())
	})
	t.Run("**struct", func(t *testing.T) {
		u := &model.Users{Id: 1, Name: "alice"}
		_, err := resolveDestination(&u)
		assert.ErrorIs(t, errMapDestination, err)
	})
}

func Test_resolveNullableDestination(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		_, err := resolveNullableDestination(nil)
		assert.ErrorIs(t, errMapRowSerialDestination, err)
	})
	t.Run("non-pointer", func(t *testing.T) {
		_, err := resolveNullableDestination(1)
		assert.ErrorIs(t, errMapRowSerialDestination, err)
	})
	t.Run("pointer of non-struct", func(t *testing.T) {
		var i int
		_, err := resolveNullableDestination(&i)
		assert.ErrorIs(t, errMapRowSerialDestination, err)
	})
	t.Run("*struct", func(t *testing.T) {
		u := &model.Users{Id: 1, Name: "alice"}
		v, err := resolveNullableDestination(u)
		assert.NoError(t, err)
		assert.Equal(t, reflect.Struct, v.Kind())
		assert.Equal(t, int64(1), v.FieldByName("Id").Int())
		assert.Equal(t, "alice", v.FieldByName("Name").String())
	})
	t.Run("**struct(nil)", func(t *testing.T) {
		var u *model.Users
		v, err := resolveNullableDestination(&u)
		assert.NoError(t, err)
		assert.Equal(t, reflect.Pointer, v.Kind())
		assert.True(t, v.IsNil())
	})
	t.Run("**struct(non-nil)", func(t *testing.T) {
		u := &model.Users{Id: 1, Name: "alice"}
		_, err := resolveNullableDestination(&u)
		assert.ErrorIs(t, errMapRowSerialDestination, err)
	})
}

func Test_resolveDestinationMany(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		_, _, err := resolveDestinationMany(nil)
		assert.ErrorIs(t, errMapManyDestination, err)
	})
	t.Run("non-pointer", func(t *testing.T) {
		_, _, err := resolveDestinationMany([]*model.Users{})
		assert.ErrorIs(t, errMapManyDestination, err)
	})
	t.Run("pointer of non-slice", func(t *testing.T) {
		var u model.Users
		_, _, err := resolveDestinationMany(&u)
		assert.ErrorIs(t, errMapManyDestination, err)
	})
	t.Run("pointer of slice of non-pointer", func(t *testing.T) {
		var users []model.Users
		_, _, err := resolveDestinationMany(&users)
		assert.ErrorIs(t, errMapManyDestination, err)
	})
	t.Run("pointer of slice of pointer of non-struct", func(t *testing.T) {
		var values []*int
		_, _, err := resolveDestinationMany(&values)
		assert.ErrorIs(t, errMapManyDestination, err)
	})
	t.Run("pointer of slice of pointer of struct", func(t *testing.T) {
		users := []*model.Users{{Id: 1, Name: "alice"}}
		typ, v, err := resolveDestinationMany(&users)
		assert.NoError(t, err)
		assert.Equal(t, reflect.TypeFor[model.Users](), typ)
		assert.Equal(t, reflect.Pointer, v.Kind())
		assert.Equal(t, reflect.Slice, v.Elem().Kind())
		assert.Len(t, v.Elem().Interface(), 1)
		assert.Equal(t, "alice", v.Elem().Index(0).Elem().FieldByName("Name").String())
	})
}

func Test_resolveDestType(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		v := reflect.ValueOf(model.Users{Id: 1, Name: "alice"})
		typ, err := resolveDestType(&v)
		assert.NoError(t, err)
		assert.Equal(t, reflect.TypeFor[model.Users](), typ)
	})
	t.Run("*struct(non-nil)", func(t *testing.T) {
		u := &model.Users{Id: 1, Name: "alice"}
		v := reflect.ValueOf(u)
		typ, err := resolveDestType(&v)
		assert.NoError(t, err)
		assert.Equal(t, reflect.TypeFor[model.Users](), typ)
	})
	t.Run("*struct(nil)", func(t *testing.T) {
		var u *model.Users
		v := reflect.ValueOf(u)
		typ, err := resolveDestType(&v)
		assert.NoError(t, err)
		assert.Equal(t, reflect.TypeFor[model.Users](), typ)
	})
	t.Run("zero value", func(t *testing.T) {
		v := reflect.Value{}
		_, err := resolveDestType(&v)
		assert.ErrorIs(t, errModelNil, err)
	})
}

func TestReflectorGetSchema(t *testing.T) {
	t.Run("returns schema for model pointer", func(t *testing.T) {
		r := &reflector{}

		schema, err := r.GetSchema(&model.Users{})

		assert.NoError(t, err)
		if !assert.NotNil(t, schema) {
			return
		}
		assert.NotNil(t, schema.autoIncrementField)
		assert.Equal(t, []int{0}, schema.primaryKeyFields)
		assert.Equal(t, []int{1, 2}, schema.updatableFields)
	})

	t.Run("returns validation error for invalid destination", func(t *testing.T) {
		r := &reflector{}

		schema, err := r.GetSchema(model.Users{})

		assert.Nil(t, schema)
		assert.ErrorIs(t, err, errMapDestination)
	})

	t.Run("uses cached schema when cache is enabled", func(t *testing.T) {
		r := &reflector{}

		s1, err := r.GetSchema(&model.Users{})
		assert.NoError(t, err)

		s2, err := r.GetSchema(&model.Users{})
		assert.NoError(t, err)
		assert.Same(t, s1, s2)
	})

	t.Run("rebuilds schema when cache is disabled", func(t *testing.T) {
		r := &reflector{noCache: true}

		s1, err := r.GetSchema(&model.Users{})
		assert.NoError(t, err)

		s2, err := r.GetSchema(&model.Users{})
		assert.NoError(t, err)
		assert.NotSame(t, s1, s2)
	})

	t.Run("is safe for concurrent access", func(t *testing.T) {
		r := &reflector{}
		const goroutines = 64
		expected, err := r.GetSchema(&model.Users{})
		assert.NoError(t, err)

		results := make(chan *modelSchema, goroutines)
		errs := make(chan error, goroutines)
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for range goroutines {
			go func() {
				defer wg.Done()
				schema, err := r.GetSchema(&model.Users{})
				if err != nil {
					errs <- err
					return
				}
				results <- schema
			}()
		}

		wg.Wait()
		close(results)
		close(errs)

		for err := range errs {
			assert.NoError(t, err)
		}

		var first *modelSchema
		count := 0
		for schema := range results {
			if !assert.NotNil(t, schema) {
				continue
			}
			assert.Same(t, expected, schema)
			first = schema
			count++
		}
		assert.Same(t, expected, first)
		assert.Equal(t, goroutines, count)
	})
}

func TestReflectorGetSchemaFromValue(t *testing.T) {
	t.Run("supports nil struct pointer value", func(t *testing.T) {
		r := &reflector{}
		var user *model.Users
		v, err := resolveNullableDestination(&user)
		assert.NoError(t, err)

		schema, err := r.GetSchemaFromValue(v)

		assert.NoError(t, err)
		assert.NotNil(t, schema)
	})

	t.Run("returns error for invalid value", func(t *testing.T) {
		r := &reflector{}
		v := reflect.Value{}

		schema, err := r.GetSchemaFromValue(&v)

		assert.Nil(t, schema)
		assert.ErrorIs(t, err, errModelNil)
	})
}

func TestReflectorClearSchemaCache(t *testing.T) {
	r := &reflector{}

	s1, err := r.GetSchema(&model.Users{})
	assert.NoError(t, err)

	r.ClearSchemaCache()

	s2, err := r.GetSchema(&model.Users{})
	assert.NoError(t, err)
	assert.NotSame(t, s1, s2)
}

func Test_typeKey(t *testing.T) {
	assert.Equal(t, "github.com/loilo-inc/exql/v3/model.Users", typeKey(reflect.TypeFor[model.Users]()))
}
