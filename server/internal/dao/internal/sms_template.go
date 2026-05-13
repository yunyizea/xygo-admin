// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SmsTemplateDao is the data access object for the table xy_sms_template.
type SmsTemplateDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  SmsTemplateColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// SmsTemplateColumns defines and stores column names for the table xy_sms_template.
type SmsTemplateColumns struct {
	Id                 string // 主键
	Title              string // 模板标题
	Code               string // 模板唯一标识
	Content            string // 短信文案
	ProviderTemplateId string // 服务商模板ID
	Variables          string // 模板变量列表 JSON
	RelatedVariableId  string // 关联文案变量ID
	Status             string // 状态：1=启用 0=禁用
	Sort               string // 排序
	Remark             string // 备注
	CreatedBy          string // 创建人ID
	UpdatedBy          string // 更新人ID
	CreateTime         string // 创建时间
	UpdateTime         string // 更新时间
}

// smsTemplateColumns holds the columns for the table xy_sms_template.
var smsTemplateColumns = SmsTemplateColumns{
	Id:                 "id",
	Title:              "title",
	Code:               "code",
	Content:            "content",
	ProviderTemplateId: "provider_template_id",
	Variables:          "variables",
	RelatedVariableId:  "related_variable_id",
	Status:             "status",
	Sort:               "sort",
	Remark:             "remark",
	CreatedBy:          "created_by",
	UpdatedBy:          "updated_by",
	CreateTime:         "create_time",
	UpdateTime:         "update_time",
}

// NewSmsTemplateDao creates and returns a new DAO object for table data access.
func NewSmsTemplateDao(handlers ...gdb.ModelHandler) *SmsTemplateDao {
	return &SmsTemplateDao{
		group:    "default",
		table:    "xy_sms_template",
		columns:  smsTemplateColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SmsTemplateDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SmsTemplateDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SmsTemplateDao) Columns() SmsTemplateColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SmsTemplateDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO.
func (dao *SmsTemplateDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
func (dao *SmsTemplateDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
