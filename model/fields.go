// This file is generated by exql. DO NOT edit.
package model

import "github.com/volatiletech/null"
import "time"

type Fields struct {
	TinyintField                   int64        `exql:"column:tinyint_field;type:tinyint(4);not null" json:"tinyint_field"`
	TinyintUnsignedField           int64        `exql:"column:tinyint_unsigned_field;type:tinyint(4) unsigned;not null" json:"tinyint_unsigned_field"`
	TinyintNullableField           null.Int64   `exql:"column:tinyint_nullable_field;type:tinyint(4)" json:"tinyint_nullable_field"`
	TinyintUnsignedNullableField   null.Int64   `exql:"column:tinyint_unsigned_nullable_field;type:tinyint(4) unsigned" json:"tinyint_unsigned_nullable_field"`
	SmallintField                  int64        `exql:"column:smallint_field;type:smallint(6);not null" json:"smallint_field"`
	SmallintUnsignedField          int64        `exql:"column:smallint_unsigned_field;type:smallint(6) unsigned;not null" json:"smallint_unsigned_field"`
	SmallintNullableField          null.Int64   `exql:"column:smallint_nullable_field;type:smallint(6)" json:"smallint_nullable_field"`
	SmallintUnsignedNullableField  null.Int64   `exql:"column:smallint_unsigned_nullable_field;type:smallint(6) unsigned" json:"smallint_unsigned_nullable_field"`
	MediumintField                 int64        `exql:"column:mediumint_field;type:mediumint(6);not null" json:"mediumint_field"`
	MediumintUnsignedField         int64        `exql:"column:mediumint_unsigned_field;type:mediumint(6) unsigned;not null" json:"mediumint_unsigned_field"`
	MediumintNullableField         null.Int64   `exql:"column:mediumint_nullable_field;type:mediumint(6)" json:"mediumint_nullable_field"`
	MediumintUnsignedNullableField null.Int64   `exql:"column:mediumint_unsigned_nullable_field;type:mediumint(6) unsigned" json:"mediumint_unsigned_nullable_field"`
	IntField                       int64        `exql:"column:int_field;type:int(11);not null" json:"int_field"`
	IntUnsignedField               int64        `exql:"column:int_unsigned_field;type:int(11) unsigned;not null" json:"int_unsigned_field"`
	IntNullableField               null.Int64   `exql:"column:int_nullable_field;type:int(11)" json:"int_nullable_field"`
	IntUnsignedNullableField       null.Int64   `exql:"column:int_unsigned_nullable_field;type:int(11) unsigned" json:"int_unsigned_nullable_field"`
	BigintField                    int64        `exql:"column:bigint_field;type:bigint(20);not null" json:"bigint_field"`
	BigintUnsignedField            uint64       `exql:"column:bigint_unsigned_field;type:bigint(20) unsigned;not null" json:"bigint_unsigned_field"`
	BigintNullableField            null.Int64   `exql:"column:bigint_nullable_field;type:bigint(20)" json:"bigint_nullable_field"`
	BigintUnsignedNullableField    null.Uint64  `exql:"column:bigint_unsigned_nullable_field;type:bigint(20) unsigned" json:"bigint_unsigned_nullable_field"`
	FloatField                     float32      `exql:"column:float_field;type:float;not null" json:"float_field"`
	FloatNullField                 null.Float32 `exql:"column:float_null_field;type:float" json:"float_null_field"`
	DoubleField                    float64      `exql:"column:double_field;type:double;not null" json:"double_field"`
	DoubleNullField                null.Float64 `exql:"column:double_null_field;type:double" json:"double_null_field"`
	TinytextField                  string       `exql:"column:tinytext_field;type:tinytext;not null" json:"tinytext_field"`
	TinytextNullField              null.String  `exql:"column:tinytext_null_field;type:tinytext" json:"tinytext_null_field"`
	MediumtextField                string       `exql:"column:mediumtext_field;type:mediumtext;not null" json:"mediumtext_field"`
	MediumtextNullField            null.String  `exql:"column:mediumtext_null_field;type:mediumtext" json:"mediumtext_null_field"`
	TextField                      string       `exql:"column:text_field;type:text;not null" json:"text_field"`
	TextNullField                  null.String  `exql:"column:text_null_field;type:text" json:"text_null_field"`
	LongtextField                  string       `exql:"column:longtext_field;type:longtext;not null" json:"longtext_field"`
	LongtextNullField              null.String  `exql:"column:longtext_null_field;type:longtext" json:"longtext_null_field"`
	VarcharFiledField              string       `exql:"column:varchar_filed_field;type:varchar(255);not null" json:"varchar_filed_field"`
	VarcharNullField               null.String  `exql:"column:varchar_null_field;type:varchar(255)" json:"varchar_null_field"`
	CharFiledField                 string       `exql:"column:char_filed_field;type:char(10);not null" json:"char_filed_field"`
	CharFiledNullField             null.String  `exql:"column:char_filed_null_field;type:char(10)" json:"char_filed_null_field"`
	DateField                      time.Time    `exql:"column:date_field;type:date;not null" json:"date_field"`
	DateNullField                  null.Time    `exql:"column:date_null_field;type:date" json:"date_null_field"`
	DatetimeField                  time.Time    `exql:"column:datetime_field;type:datetime;not null" json:"datetime_field"`
	DatetimeNullField              null.Time    `exql:"column:datetime_null_field;type:datetime" json:"datetime_null_field"`
	TimeField                      time.Time    `exql:"column:time_field;type:time;not null" json:"time_field"`
	TimeNullField                  null.Time    `exql:"column:time_null_field;type:time" json:"time_null_field"`
	TimestampField                 time.Time    `exql:"column:timestamp_field;type:timestamp;not null;on;update;CURRENT_TIMESTAMP" json:"timestamp_field"`
	TimestampNullField             time.Time    `exql:"column:timestamp_null_field;type:timestamp;not null" json:"timestamp_null_field"`
	TinyblobField                  []byte       `exql:"column:tinyblob_field;type:tinyblob;not null" json:"tinyblob_field"`
	TinyblobNullField              null.Bytes   `exql:"column:tinyblob_null_field;type:tinyblob" json:"tinyblob_null_field"`
	MediumblobField                []byte       `exql:"column:mediumblob_field;type:mediumblob;not null" json:"mediumblob_field"`
	MediumblobNullField            null.Bytes   `exql:"column:mediumblob_null_field;type:mediumblob" json:"mediumblob_null_field"`
	BlobField                      []byte       `exql:"column:blob_field;type:blob;not null" json:"blob_field"`
	BlobNullField                  null.Bytes   `exql:"column:blob_null_field;type:blob" json:"blob_null_field"`
	LongblobField                  []byte       `exql:"column:longblob_field;type:longblob;not null" json:"longblob_field"`
	LongblobNullField              null.Bytes   `exql:"column:longblob_null_field;type:longblob" json:"longblob_null_field"`
}

func (f *Fields) TableName() string {
	return "fields"
}

type fieldsTable struct {
}

var FieldsTable = &fieldsTable{}

func (f *fieldsTable) Name() string {
	return "fields"
}
