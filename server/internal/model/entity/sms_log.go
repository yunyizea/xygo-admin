// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/encoding/gjson"
)

// SmsLog is the golang structure for table sms_log.
type SmsLog struct {
	Id           uint64      `json:"id"           orm:"id"            description:"主键"`
	Phone        string      `json:"phone"        orm:"phone"         description:"手机号"`
	TemplateCode string      `json:"templateCode" orm:"template_code" description:"使用的模板标识"`
	Driver       string      `json:"driver"       orm:"driver"        description:"驱动名"`
	Content      string      `json:"content"      orm:"content"       description:"实际发送内容"`
	Params       *gjson.Json `json:"params"       orm:"params"        description:"发送参数 JSON"`
	Status       int         `json:"status"       orm:"status"        description:"状态：1=成功 0=失败"`
	RequestId    string      `json:"requestId"    orm:"request_id"    description:"服务商返回请求ID"`
	ErrorMsg     string      `json:"errorMsg"     orm:"error_msg"     description:"错误信息"`
	CreateTime   uint64      `json:"createTime"   orm:"create_time"   description:"发送时间"`
}
