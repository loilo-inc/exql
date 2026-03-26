package exql

import (
	"fmt"
	"reflect"

	q "github.com/loilo-inc/exql/v3/query"
)

func Where(str string, args ...any) q.Condition {
	return q.Cond(str, args...)
}

func QueryForInsert(modelPtr Model) (q.Query, *reflect.Value, error) {
	dest, err := resolveDestination(modelPtr)
	if err != nil {
		return nil, nil, err
	}
	ms, err := parseUpsertSchema(dest.Type(), false)
	if err != nil {
		return nil, nil, err
	}
	tableName := modelPtr.TableName()
	if tableName == "" {
		return nil, nil, errTableNameEmpty
	}
	v, err := ms.aggregateValue(modelPtr)
	if err != nil {
		return nil, nil, err
	}
	b := q.NewBuilder()
	iter := q.NewKeyIterator(v.values)
	cols := q.Cols(iter.Keys()...)
	vals := q.Vals(iter.Values())
	b.Sprintf("INSERT INTO `%s`", tableName)
	b.Query("(:?) VALUES (:?)", cols, vals)
	return b.Build(), v.autoIncrementField, nil
}

func QueryForBulkInsert[T Model](modelPtrs ...T) (q.Query, error) {
	if len(modelPtrs) == 0 {
		return nil, fmt.Errorf("empty list")
	}
	destType, err := resolveDestination(modelPtrs[0])
	if err != nil {
		return nil, err
	}
	ms, err := parseUpsertSchema(destType.Type(), false)
	if err != nil {
		return nil, err
	}
	tableName := modelPtrs[0].TableName()
	if tableName == "" {
		return nil, errTableNameEmpty
	}
	var cols q.Query
	b := q.NewBuilder()
	vals := q.NewBuilder()
	for _, v := range modelPtrs {
		data, err := ms.aggregateValue(v)
		if err != nil {
			return nil, err
		}
		iter := q.NewKeyIterator(data.values)
		if cols == nil {
			cols = q.Cols(iter.Keys()...)
		}
		vals.Query("(:?)", q.Vals(iter.Values()))
	}
	b.Sprintf("INSERT INTO `%s`", tableName)
	b.Query("(:?) VALUES :?", cols, vals.Join(","))
	return b.Build(), nil
}

func QueryForUpdateModel(
	updateStructPtr ModelUpdate,
	where q.Condition,
) (q.Query, error) {
	if updateStructPtr == nil {
		return nil, errModelNil
	}
	dest, err := resolveDestination(updateStructPtr)
	if err != nil {
		return nil, err
	}
	ms, err := parseUpsertSchema(dest.Type(), true)
	if err != nil {
		return nil, err
	}
	tableName := updateStructPtr.UpdateTableName()
	if tableName == "" {
		return nil, errTableNameEmpty
	}
	v, err := ms.aggregateValue(updateStructPtr)
	if err != nil {
		return nil, err
	}
	if len(v.values) == 0 {
		return nil, fmt.Errorf("no updatable fields with non-nil value")
	}
	b := q.NewBuilder()
	b.Sprintf("UPDATE `%s`", tableName)
	b.Query("SET :? WHERE :?", q.Set(v.values), where)
	return b.Build(), nil
}
