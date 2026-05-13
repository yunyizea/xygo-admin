// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SmsLogDao is the data access object for the table xy_sms_log.
type SmsLogDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SmsLogColumns      // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SmsLogColumns defines and stores column names for the table xy_sms_log.
type SmsLogColumns struct {
	Id           string // 主键
	Phone        string // 手机号
	TemplateCode string // 使用的模板标识
	Driver       string // 驱动名
	Content      string // 实际发送内容
	Params       string // 发送参数 JSON
	Status       string // 状态：1=成功 0=失败
	RequestId    string // 服务商返回请求ID
	ErrorMsg     string // 错误信息
	CreateTime   string // 发送时间
}

// smsLogColumns holds the columns for the table xy_sms_log.
var smsLogColumns = SmsLogColumns{
	Id:           "id",
	Phone:        "phone",
	TemplateCode: "template_code",
	Driver:       "driver",
	Content:      "content",
	Params:       "params",
	Status:       "status",
	RequestId:    "request_id",
	ErrorMsg:     "error_msg",
	CreateTime:   "create_time",
}

// NewSmsLogDao creates and returns a new DAO object for table data access.
func NewSmsLogDao(handlers ...gdb.ModelHandler) *SmsLogDao {
	return &SmsLogDao{
		group:    "default",
		table:    "xy_sms_log",
		columns:  smsLogColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SmsLogDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SmsLogDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SmsLogDao) Columns() SmsLogColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SmsLogDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO.
func (dao *SmsLogDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SmsLogDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
