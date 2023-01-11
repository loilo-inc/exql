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
}

func (*NoTag) TableName() string {
	return "dummy"
}

type BadTableName struct {
	Id int `exql:"column:id;primary;auto_increment"`
}

func (BadTableName) TableName() string {
	return ""
}

type NoPrimaryKey struct {
	Id int `exql:"column:id;auto_increment"`
}

func (NoPrimaryKey) TableName() string {
	return ""
}

type NoColumnTag struct {
	Id int `exql:"primary;auto_increment"`
}

func (NoColumnTag) TableName() string {
	return ""
}

type BadTag struct {
	Id int `exql:"a;a:1"`
}

func (BadTag) TableName() string {
	return ""
}

type NoAutoIncrementKey struct {
	Id   int    `exql:"column:id;primary"`
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
