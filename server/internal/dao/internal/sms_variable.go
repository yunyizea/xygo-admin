// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SmsVariableDao is the data access object for the table xy_sms_variable.
type SmsVariableDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SmsVariableColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SmsVariableColumns defines and stores column names for the table xy_sms_variable.
type SmsVariableColumns struct {
	Id          string // 主键
	Title       string // 变量标题
	Name        string // 变量名
	SourceType  string // 来源类型
	SqlQuery    string // SQL查询语句
	MethodName  string // Helper方法路径
	SharedCount string // 共通数据数
	Status      string // 状态：1=启用 0=禁用
	CreatedBy   string // 创建人ID
	UpdatedBy   string // 更新人ID
	CreateTime  string // 创建时间
	UpdateTime  string // 更新时间
}

// smsVariableColumns holds the columns for the table xy_sms_variable.
var smsVariableColumns = SmsVariableColumns{
	Id:          "id",
	Title:       "title",
	Name:        "name",
	SourceType:  "source_type",
	SqlQuery:    "sql_query",
	MethodName:  "method_name",
	SharedCount: "shared_count",
	Status:      "status",
	CreatedBy:   "created_by",
	UpdatedBy:   "updated_by",
	CreateTime:  "create_time",
	UpdateTime:  "update_time",
}

// NewSmsVariableDao creates and returns a new DAO object for table data access.
func NewSmsVariableDao(handlers ...gdb.ModelHandler) *SmsVariableDao {
	return &SmsVariableDao{
		group:    "default",
		table:    "xy_sms_variable",
		columns:  smsVariableColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SmsVariableDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SmsVariableDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SmsVariableDao) Columns() SmsVariableColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SmsVariableDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO.
func (dao *SmsVariableDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SmsVariableDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
