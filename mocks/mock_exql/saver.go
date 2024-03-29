// Code generated by MockGen. DO NOT EDIT.
// Source: saver.go

// Package mock_exql is a generated GoMock package.
package mock_exql

import (
	context "context"
	sql "database/sql"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	exql "github.com/loilo-inc/exql/v2"
	query "github.com/loilo-inc/exql/v2/query"
)

// MockSaver is a mock of Saver interface.
type MockSaver struct {
	ctrl     *gomock.Controller
	recorder *MockSaverMockRecorder
}

// MockSaverMockRecorder is the mock recorder for MockSaver.
type MockSaverMockRecorder struct {
	mock *MockSaver
}

// NewMockSaver creates a new mock instance.
func NewMockSaver(ctrl *gomock.Controller) *MockSaver {
	mock := &MockSaver{ctrl: ctrl}
	mock.recorder = &MockSaverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSaver) EXPECT() *MockSaverMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockSaver) Delete(table string, where query.Condition) (sql.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", table, where)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Delete indicates an expected call of Delete.
func (mr *MockSaverMockRecorder) Delete(table, where interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockSaver)(nil).Delete), table, where)
}

// DeleteContext mocks base method.
func (m *MockSaver) DeleteContext(ctx context.Context, table string, where query.Condition) (sql.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteContext", ctx, table, where)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteContext indicates an expected call of DeleteContext.
func (mr *MockSaverMockRecorder) DeleteContext(ctx, table, where interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteContext", reflect.TypeOf((*MockSaver)(nil).DeleteContext), ctx, table, where)
}

// Exec mocks base method.
func (m *MockSaver) Exec(query query.Query) (sql.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exec", query)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec.
func (mr *MockSaverMockRecorder) Exec(query interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockSaver)(nil).Exec), query)
}

// ExecContext mocks base method.
func (m *MockSaver) ExecContext(ctx context.Context, query query.Query) (sql.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExecContext", ctx, query)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExecContext indicates an expected call of ExecContext.
func (mr *MockSaverMockRecorder) ExecContext(ctx, query interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExecContext", reflect.TypeOf((*MockSaver)(nil).ExecContext), ctx, query)
}

// Insert mocks base method.
func (m *MockSaver) Insert(structPtr exql.Model) (sql.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", structPtr)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Insert indicates an expected call of Insert.
func (mr *MockSaverMockRecorder) Insert(structPtr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*MockSaver)(nil).Insert), structPtr)
}

// InsertContext mocks base method.
func (m *MockSaver) InsertContext(ctx context.Context, structPtr exql.Model) (sql.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertContext", ctx, structPtr)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertContext indicates an expected call of InsertContext.
func (mr *MockSaverMockRecorder) InsertContext(ctx, structPtr interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertContext", reflect.TypeOf((*MockSaver)(nil).InsertContext), ctx, structPtr)
}

// Query mocks base method.
func (m *MockSaver) Query(query query.Query) (*sql.Rows, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Query", query)
	ret0, _ := ret[0].(*sql.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query.
func (mr *MockSaverMockRecorder) Query(query interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockSaver)(nil).Query), query)
}

// QueryContext mocks base method.
func (m *MockSaver) QueryContext(ctx context.Context, query query.Query) (*sql.Rows, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryContext", ctx, query)
	ret0, _ := ret[0].(*sql.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryContext indicates an expected call of QueryContext.
func (mr *MockSaverMockRecorder) QueryContext(ctx, query interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryContext", reflect.TypeOf((*MockSaver)(nil).QueryContext), ctx, query)
}

// QueryRow mocks base method.
func (m *MockSaver) QueryRow(query query.Query) (*sql.Row, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryRow", query)
	ret0, _ := ret[0].(*sql.Row)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryRow indicates an expected call of QueryRow.
func (mr *MockSaverMockRecorder) QueryRow(query interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRow", reflect.TypeOf((*MockSaver)(nil).QueryRow), query)
}

// QueryRowContext mocks base method.
func (m *MockSaver) QueryRowContext(ctx context.Context, query query.Query) (*sql.Row, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryRowContext", ctx, query)
	ret0, _ := ret[0].(*sql.Row)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryRowContext indicates an expected call of QueryRowContext.
func (mr *MockSaverMockRecorder) QueryRowContext(ctx, query interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryRowContext", reflect.TypeOf((*MockSaver)(nil).QueryRowContext), ctx, query)
}

// Update mocks base method.
func (m *MockSaver) Update(table string, set map[string]any, where query.Condition) (sql.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", table, set, where)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockSaverMockRecorder) Update(table, set, where interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockSaver)(nil).Update), table, set, where)
}

// UpdateContext mocks base method.
func (m *MockSaver) UpdateContext(ctx context.Context, table string, set map[string]any, where query.Condition) (sql.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateContext", ctx, table, set, where)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateContext indicates an expected call of UpdateContext.
func (mr *MockSaverMockRecorder) UpdateContext(ctx, table, set, where interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateContext", reflect.TypeOf((*MockSaver)(nil).UpdateContext), ctx, table, set, where)
}

// UpdateModel mocks base method.
func (m *MockSaver) UpdateModel(updaterStructPtr exql.ModelUpdate, where query.Condition) (sql.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateModel", updaterStructPtr, where)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateModel indicates an expected call of UpdateModel.
func (mr *MockSaverMockRecorder) UpdateModel(updaterStructPtr, where interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateModel", reflect.TypeOf((*MockSaver)(nil).UpdateModel), updaterStructPtr, where)
}

// UpdateModelContext mocks base method.
func (m *MockSaver) UpdateModelContext(ctx context.Context, updaterStructPtr exql.ModelUpdate, where query.Condition) (sql.Result, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateModelContext", ctx, updaterStructPtr, where)
	ret0, _ := ret[0].(sql.Result)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateModelContext indicates an expected call of UpdateModelContext.
func (mr *MockSaverMockRecorder) UpdateModelContext(ctx, updaterStructPtr, where interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateModelContext", reflect.TypeOf((*MockSaver)(nil).UpdateModelContext), ctx, updaterStructPtr, where)
}
