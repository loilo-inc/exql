// This file is generated by exql. DO NOT edit.
package model

type GroupUsers struct {
	Id      int64 `exql:"column:id;type:int(11);primary;not null;auto_increment" json:"id"`
	UserId  int64 `exql:"column:user_id;type:int(11);not null" json:"user_id"`
	GroupId int64 `exql:"column:group_id;type:int(11);not null" json:"group_id"`
}

func (g *GroupUsers) TableName() string {
	return "group_users"
}

type groupUsersTable struct {
}

var GroupUsersTable = &groupUsersTable{}

func (g *groupUsersTable) Name() string {
	return "group_users"
}
