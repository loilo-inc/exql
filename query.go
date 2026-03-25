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
	ms, err := aggregateUpsertSchema(dest.Type(), false)
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
	cols := q.Cols(v.values.Keys()...)
	vals := q.Vals(v.values.Values())
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
	ms, err := aggregateUpsertSchema(destType.Type(), false)
	if err != nil {
		return nil, err
	}
	tableName := modelPtrs[0].TableName()
	if tableName == "" {
		return nil, errTableNameEmpty
	}
	var head *modelValue
	b := q.NewBuilder()
	vals := q.NewBuilder()
	for _, v := range modelPtrs {
		if data, err := ms.aggregateValue(v); err != nil {
			return nil, err
		} else {
			if head == nil {
				head = data
			}
			vals.Query("(:?)", q.Vals(data.values.Values()))
		}
	}
	b.Sprintf("INSERT INTO `%s`", tableName)
	b.Query("(:?) VALUES :?", q.Cols(head.values.Keys()...), vals.Join(","))
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
	ms, err := aggregateUpsertSchema(dest.Type(), true)
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
	if v.values.Size() == 0 {
		return nil, fmt.Errorf("no updatable fields with non-nil value")
	}
	b := q.NewBuilder()
	b.Sprintf("UPDATE `%s`", tableName)
	b.Query("SET :? WHERE :?", q.Set(v.values.Map()), where)
	return b.Build(), nil
}
