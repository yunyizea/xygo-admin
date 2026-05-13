package admin

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ===================== 短信发送日志 =====================

type SmsLogListReq struct {
	g.Meta       `path:"/admin/sms/log/list" method:"get" tags:"AdminSms" summary:"短信发送日志列表"`
	Page         int    `p:"page" json:"page" d:"1"`
	Size         int    `p:"size" json:"size" d:"20"`
	Phone        string `p:"phone" json:"phone" dc:"手机号"`
	TemplateCode string `p:"templateCode" json:"templateCode" dc:"模板标识"`
	Status       int    `p:"status" json:"status" d:"-1" dc:"状态：-1=全部 0=失败 1=成功"`
	Driver       string `p:"driver" json:"driver" dc:"驱动名"`
}

type SmsLogListItem struct {
	Id           uint64      `json:"id"`
	Phone        string      `json:"phone"`
	TemplateCode string      `json:"templateCode"`
	Driver       string      `json:"driver"`
	Content      string      `json:"content"`
	Params       interface{} `json:"params"`
	Status       int         `json:"status"`
	RequestId    string      `json:"requestId"`
	ErrorMsg     string      `json:"errorMsg"`
	CreateTime   uint64      `json:"createTime"`
}

type SmsLogListRes struct {
	List  []SmsLogListItem `json:"list"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
}
