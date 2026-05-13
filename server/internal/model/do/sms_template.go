// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// SmsTemplate is the golang structure of table xy_sms_template for DAO operations like Where/Data.
type SmsTemplate struct {
	g.Meta             `orm:"table:xy_sms_template, do:true"`
	Id                 any         // 主键
	Title              any         // 模板标题
	Code               any         // 模板唯一标识
	Content            any         // 短信文案
	ProviderTemplateId any         // 服务商模板ID
	Variables          *gjson.Json // 模板变量列表 JSON
	RelatedVariableId  any         // 关联文案变量ID
	Status             any         // 状态：1=启用 0=禁用
	Sort               any         // 排序
	Remark             any         // 备注
	CreatedBy          any         // 创建人ID
	UpdatedBy          any         // 更新人ID
	CreateTime         any         // 创建时间
	UpdateTime         any         // 更新时间
}
