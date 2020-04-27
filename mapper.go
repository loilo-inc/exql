package exql

import (
	"database/sql"
	"fmt"
	"github.com/apex/log"
	"reflect"
	"strings"
)

func (d *db) MapRows(rows *sql.Rows, structPtrOrSlicePtr interface{}) error {
	destValue := reflect.ValueOf(structPtrOrSlicePtr)
	destType := destValue.Type()
	if destType.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be pointer of struct or slice of struct")
	}
	destType = destType.Elem()
	if destType.Kind() == reflect.Slice {
		// []*Model -> *Model
		sliceType := destType.Elem()
		if sliceType.Kind() != reflect.Ptr {
			return fmt.Errorf("slice type is not pointer")
		}
		// *Model -> Model
		sliceType = sliceType.Elem()
		log.Infof("%s", destType.Name())
		for rows.Next() {
			// modelValue := SliceType{}
			modelValue := reflect.New(sliceType).Elem()
			if err := d.mapRow(rows, &modelValue); err != nil {
				return err
			}
			// *dest = append(*dest, i)
			destValue.Elem().Set(reflect.Append(destValue.Elem(), modelValue.Addr()))
		}
		return nil
	} else if destType.Kind() == reflect.Struct {
		if rows.Next() {
			destValue = destValue.Elem()
			return d.mapRow(rows, &destValue)
		} else {
			return fmt.Errorf("rows is empty")
		}
	} else {
		return fmt.Errorf("unsupported type")
	}
}

func ParseTags(tag string) map[string]string {
	tags := strings.Split(tag, ";")
	ret := make(map[string]string)
	for _, tag := range tags {
		kv := strings.Split(tag, ":")
		if len(kv) == 0 {
			continue
		}
		if len(kv) == 1 {
			ret[kv[0]] = ""
		} else if len(kv) == 2 {
			ret[kv[0]] = kv[1]
		} else {
			panic(fmt.Sprintf("invalid tag format: %s", tag))
		}
	}
	return ret
}

func (d *db) mapRow(row *sql.Rows, dest *reflect.Value) error {
	// *Model
	destType := dest.Type()
	fields := make(map[string]int)
	for i := 0; i < destType.NumField(); i++ {
		f := destType.Field(i)
		tag := f.Tag.Get("exql")
		if tag != "" {
			if f.Type.Kind() == reflect.Ptr {
				return fmt.Errorf("struct field must not be pointer: %s %s", f.Type.Name(), f.Type.Kind())
			}
			tags := ParseTags(tag)
			col := tags["column"]
			fields[col] = i
		}
	}
	cols, err := row.ColumnTypes()
	if err != nil {
		return err
	}
	destVals := make([]interface{}, len(cols))
	for j, col := range cols {
		if fIndex, ok := fields[col.Name()]; ok {
			f := dest.Field(fIndex)
			destVals[j] = f.Addr().Interface()
		} else {
			ns := &nullScanner{}
			destVals[j] = ns
		}
	}
	return row.Scan(destVals...)
}

type nullScanner struct {
}

func (n *nullScanner) Scan(_ interface{}) error {
	// noop
	return nil
}
