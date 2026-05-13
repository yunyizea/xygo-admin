// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/encoding/gjson"
)

// SmsTemplate is the golang structure for table sms_template.
type SmsTemplate struct {
	Id                 uint64      `json:"id"                 orm:"id"                   description:"主键"`
	Title              string      `json:"title"              orm:"title"                description:"模板标题"`
	Code               string      `json:"code"               orm:"code"                 description:"模板唯一标识"`
	Content            string      `json:"content"            orm:"content"              description:"短信文案"`
	ProviderTemplateId string      `json:"providerTemplateId" orm:"provider_template_id" description:"服务商模板ID"`
	Variables          *gjson.Json `json:"variables"          orm:"variables"            description:"模板变量列表 JSON"`
	RelatedVariableId  uint64      `json:"relatedVariableId"  orm:"related_variable_id"  description:"关联文案变量ID"`
	Status             int         `json:"status"             orm:"status"               description:"状态：1=启用 0=禁用"`
	Sort               int         `json:"sort"               orm:"sort"                 description:"排序"`
	Remark             string      `json:"remark"             orm:"remark"               description:"备注"`
	CreatedBy          uint64      `json:"createdBy"          orm:"created_by"           description:"创建人ID"`
	UpdatedBy          uint64      `json:"updatedBy"          orm:"updated_by"           description:"更新人ID"`
	CreateTime         uint64      `json:"createTime"         orm:"create_time"          description:"创建时间"`
	UpdateTime         uint64      `json:"updateTime"         orm:"update_time"          description:"更新时间"`
}
