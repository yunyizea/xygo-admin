// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

// SmsLog is the golang structure of table xy_sms_log for DAO operations like Where/Data.
type SmsLog struct {
	g.Meta       `orm:"table:xy_sms_log, do:true"`
	Id           any         // 主键
	Phone        any         // 手机号
	TemplateCode any         // 使用的模板标识
	Driver       any         // 驱动名
	Content      any         // 实际发送内容
	Params       *gjson.Json // 发送参数 JSON
	Status       any         // 状态：1=成功 0=失败
	RequestId    any         // 服务商返回请求ID
	ErrorMsg     any         // 错误信息
	CreateTime   any         // 发送时间
}
