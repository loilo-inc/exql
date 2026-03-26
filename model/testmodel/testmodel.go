package testmodel

type MultiplePrimaryKey struct {
	Pk1   string `exql:"column:pk1;primary"`
	Pk2   string `exql:"column:pk2;primary"`
	Other int    `exql:"column:other"`
}

func (*MultiplePrimaryKey) TableName() string {
	return "dummy"
}

type NoTag struct {
	Id int64
}

func (*NoTag) TableName() string {
	return "dummy"
}

type BadTableName struct {
	Id int64 `exql:"column:id;primary;auto_increment"`
}

func (BadTableName) TableName() string {
	return ""
}

type NoPrimaryKey struct {
	Id int64 `exql:"column:id;auto_increment"`
}

func (NoPrimaryKey) TableName() string {
	return ""
}

type NoColumnTag struct {
	Id int64 `exql:"primary;auto_increment"`
}

func (NoColumnTag) TableName() string {
	return ""
}

type BadTag struct {
	Id int64 `exql:"a;a:1"`
}

func (BadTag) TableName() string {
	return ""
}

type NoAutoIncrementKey struct {
	Id   int64  `exql:"column:id;primary"`
	Name string `exql:"column:name"`
}

func (s *NoAutoIncrementKey) TableName() string {
	return "sampleNoAutoIncrementKey"
}

type PrimaryUint64 struct {
	Id   uint64 `exql:"column:id;primary;auto_increment"`
	Name string `exql:"column:name"`
}

func (s *PrimaryUint64) TableName() string {
	return "samplePrimaryUint64"
}

type UpdateSampleInvalidTag struct {
	Id *int `exql:"column::"`
}

func (UpdateSampleInvalidTag) UpdateTableName() string {
	return ""
}

type UpdateSampleNotPtr struct {
	Id int64 `exql:"column:id;primary"`
}

func (UpdateSampleNotPtr) UpdateTableName() string {
	return ""
}

type UpdateSample struct {
	Id *int64 `exql:"column:id;primary"`
}

func (UpdateSample) UpdateTableName() string {
	return "table"
}

type UpdateSampleNoFields struct {
}

func (UpdateSampleNoFields) UpdateTableName() string {
	return ""
}

type UpdateSampleNoColumn struct {
	Id *int `exql:"row:id"`
}

func (UpdateSampleNoColumn) UpdateTableName() string {
	return "table"
}
