package admin

import (
	"github.com/gogf/gf/v2/frame/g"

	"xygo/internal/model/input/adminin"
)

// ===================== 短信模板列表 =====================

type SmsTemplateListReq struct {
	g.Meta `path:"/admin/sms/template/list" method:"get" tags:"AdminSms" summary:"短信模板列表"`
	Page   int    `p:"page" json:"page" d:"1"`
	Size   int    `p:"size" json:"size" d:"20"`
	Status int    `p:"status" json:"status" d:"-1" dc:"状态筛选：-1=全部 0=禁用 1=启用"`
	Code   string `p:"code" json:"code" dc:"模板标识"`
	Title  string `p:"title" json:"title" dc:"模板标题关键词"`
}

type SmsTemplateListRes struct {
	List  []adminin.SmsTemplateListItem `json:"list"`
	Total int                           `json:"total"`
	Page  int                           `json:"page"`
	Size  int                           `json:"size"`
}

// ===================== 保存短信模板 =====================

type SmsTemplateSaveReq struct {
	g.Meta             `path:"/admin/sms/template/save" method:"post" tags:"AdminSms" summary:"新增/编辑短信模板"`
	Id                 uint64 `json:"id" dc:"ID，0=新增"`
	Title              string `json:"title" v:"required#模板标题必填"`
	Code               string `json:"code" v:"required#模板标识必填"`
	Content            string `json:"content"`
	ProviderTemplateId string `json:"providerTemplateId"`
	Variables          string `json:"variables" dc:"模板变量 JSON 字符串"`
	RelatedVariableId  uint64 `json:"relatedVariableId"`
	Status             int    `json:"status" d:"1"`
	Sort               int    `json:"sort"`
	Remark             string `json:"remark"`
}

type SmsTemplateSaveRes struct{}

// ===================== 删除短信模板 =====================

type SmsTemplateDeleteReq struct {
	g.Meta `path:"/admin/sms/template/delete" method:"post" tags:"AdminSms" summary:"删除短信模板"`
	Id     uint64 `json:"id" v:"required#ID必填"`
}

type SmsTemplateDeleteRes struct{}

// ===================== 测试发送 =====================

type SmsTemplateTestReq struct {
	g.Meta `path:"/admin/sms/template/test" method:"post" tags:"AdminSms" summary:"测试发送短信"`
	Id     uint64 `json:"id" v:"required#模板ID必填"`
	Phone  string `json:"phone" v:"required#手机号必填"`
}

type SmsTemplateTestRes struct {
	Success   bool   `json:"success"`
	RequestId string `json:"requestId"`
	Message   string `json:"message"`
}
