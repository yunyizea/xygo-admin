package admin

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ===================== 短信变量列表 =====================

type SmsVariableListReq struct {
	g.Meta `path:"/admin/sms/variable/list" method:"get" tags:"AdminSms" summary:"短信变量列表"`
	Page   int    `p:"page" json:"page" d:"1"`
	Size   int    `p:"size" json:"size" d:"20"`
	Name   string `p:"name" json:"name" dc:"变量名关键词"`
}

type SmsVariableListItem struct {
	Id          uint64 `json:"id"`
	Title       string `json:"title"`
	Name        string `json:"name"`
	SourceType  int    `json:"sourceType"`
	SqlQuery    string `json:"sqlQuery"`
	MethodName  string `json:"methodName"`
	SharedCount int    `json:"sharedCount"`
	Status      int    `json:"status"`
	CreateTime  uint64 `json:"createTime"`
	UpdateTime  uint64 `json:"updateTime"`
}

type SmsVariableListRes struct {
	List  []SmsVariableListItem `json:"list"`
	Total int                   `json:"total"`
	Page  int                   `json:"page"`
	Size  int                   `json:"size"`
}

// ===================== 保存短信变量 =====================

type SmsVariableSaveReq struct {
	g.Meta     `path:"/admin/sms/variable/save" method:"post" tags:"AdminSms" summary:"新增/编辑短信变量"`
	Id         uint64 `json:"id" dc:"ID，0=新增"`
	Title      string `json:"title" v:"required#变量标题必填"`
	Name       string `json:"name" v:"required#变量名必填"`
	SourceType int    `json:"sourceType" d:"1" dc:"来源类型：1=字段 2=SQL 3=Helper"`
	SqlQuery   string `json:"sqlQuery"`
	MethodName string `json:"methodName"`
	Status     int    `json:"status" d:"1"`
}

type SmsVariableSaveRes struct{}

// ===================== 删除短信变量 =====================

type SmsVariableDeleteReq struct {
	g.Meta `path:"/admin/sms/variable/delete" method:"post" tags:"AdminSms" summary:"删除短信变量"`
	Id     uint64 `json:"id" v:"required#ID必填"`
}

type SmsVariableDeleteRes struct{}
