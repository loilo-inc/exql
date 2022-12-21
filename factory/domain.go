package factory

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	_ "github.com/go-sql-driver/mysql"

	"github.com/loilo-inc/exql"
	"github.com/loilo-inc/exql/query"
	"golang.org/x/xerrors"
)

type D struct {
	db                   exql.DB
	mutex                sync.RWMutex
	models               []exql.Model
	createdFlags         map[any]struct{}
	elasticsearchIndices map[string]string
}

func Domain(t *testing.T, db exql.DB) *D {
	ret := &D{
		db:                   db,
		mutex:                sync.RWMutex{},
		models:               []exql.Model{},
		elasticsearchIndices: map[string]string{},
		createdFlags:         make(map[any]struct{}),
	}
	t.Cleanup(func() { ret.Reset() })
	return ret
}

func (d *D) isCreated(model exql.Model) bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	_, ok := d.createdFlags[model]
	return ok
}

func (d *D) store(model exql.Model) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if _, ok := d.createdFlags[model]; ok {
		return false
	}
	d.models = append(d.models, model)
	d.createdFlags[model] = struct{}{}
	return true
}

func (d *D) Create(models ...exql.Model) {
	err := d.create(models...)
	if err != nil {
		panic(err)
	}
}

func (d *D) create(models ...exql.Model) error {
	for _, model := range models {
		if d.isCreated(model) {
			return xerrors.New("model is already created by Create()")
		}
		if _, err := d.db.Insert(model); err != nil {
			return xerrors.Errorf("error in Insert: %w", err)
		}
		if !d.store(model) {
			return xerrors.New("model is already created by Create()")
		}
	}
	return nil
}

func (d *D) Save(models ...exql.Model) {
	err := d.save(models...)
	if err != nil {
		panic(err)
	}
}

func (d *D) save(models ...exql.Model) error {
	for _, v := range models {
		if !d.isCreated(v) {
			return errors.New("model is not created from Create()")
		}
		if err := d.saveModel(v); err != nil {
			return err
		}
	}
	return nil
}

func (d *D) saveModel(v exql.Model) error {
	o, err := aggregateMetadata(v)
	if err != nil {
		return err
	}
	cond := map[string]any{}
	cond[o.autoIncrementColumn] = o.autoIncrementValue
	q := query.Update{
		Table: o.tableName,
		Set:   o.fields,
		Where: exql.WhereEx(cond),
	}
	if _, err = d.db.Exec(q); err != nil {
		return xerrors.Errorf("error in Update: %w", err)
	} else {
		return nil
	}
}

func (d *D) Reset() {
	if err := d.reest(); err != nil {
		panic(err)
	}
}

func (d *D) reest() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	var firstErr error
	for i := len(d.models) - 1; i >= 0; i-- {
		v := d.models[i]
		if err := d.deleteModel(v); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	d.models = []exql.Model{}
	d.createdFlags = make(map[any]struct{})
	return firstErr
}

func (d *D) deleteModel(v exql.Model) error {
	o, err := aggregateMetadata(v)
	if err != nil {
		return err
	}
	cond := map[string]any{}
	cond[o.autoIncrementColumn] = o.autoIncrementValue
	q := query.Delete{
		From:  o.tableName,
		Where: exql.WhereEx(cond),
	}
	if _, err = d.db.Exec(q); err != nil {
		return xerrors.Errorf("error in Exec: %w", err)
	} else {
		return nil
	}
}

type tableMetadata struct {
	tableName           string
	autoIncrementColumn string
	autoIncrementValue  any
	fields              map[string]any
}

func aggregateMetadata(v exql.Model) (*tableMetadata, error) {
	val := reflect.ValueOf(v)
	valType := val.Type()
	if valType.Kind() != reflect.Ptr {
		panic("v is not pointer of struct")
	}
	// *Model -> Model
	valType = valType.Elem()
	var autoIncrementField *string
	var autoIncrementValue *reflect.Value
	fields := make(map[string]any)
	for i := 0; i < valType.NumField(); i++ {
		f := valType.Field(i)
		if tag, ok := f.Tag.Lookup("exql"); ok {
			tags, err := exql.ParseTags(tag)
			if err != nil {
				panic(err)
			}
			e := val.Elem()
			col, ok := tags["column"]
			if !ok {
				return nil, errors.New("column not found")
			}
			if _, ok := tags["auto_increment"]; ok {
				autoIncrementField = &f.Name
				autoValue := e.Field(i)
				autoIncrementValue = &autoValue
				// exclude auto_increment field
				continue
			}
			fields[col] = e.Field(i).Interface()
		}
	}
	tableName := v.TableName()
	if tableName == "" {
		return nil, errors.New("empty table name")
	}
	return &tableMetadata{
		tableName:           tableName,
		autoIncrementColumn: *autoIncrementField,
		autoIncrementValue:  autoIncrementValue.Interface(),
		fields:              fields,
	}, nil
}
