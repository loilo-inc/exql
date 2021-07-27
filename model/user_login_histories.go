// This file is generated by exql. DO NOT edit.
package model

import "time"

type UserLoginHistories struct {
	Id        int64     `exql:"column:id;type:int(11);primary;not null;auto_increment" json:"id"`
	UserId    int64     `exql:"column:user_id;type:int(11);not null" json:"user_id"`
	CreatedAt time.Time `exql:"column:created_at;type:datetime;primary;not null" json:"created_at"`
}

func (u *UserLoginHistories) TableName() string {
	return "user_login_histories"
}

type userLoginHistoriesTable struct {
}

var UserLoginHistoriesTable = &userLoginHistoriesTable{}

func (u *userLoginHistoriesTable) Name() string {
	return "user_login_histories"
}
