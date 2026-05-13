package service

import (
	"context"
	"xygo/internal/model/input/adminin"
)

type (
	ISms interface {
		// TemplateList 短信模板列表
		TemplateList(ctx context.Context, in *adminin.SmsTemplateListInp) (*adminin.SmsTemplateListModel, error)
		// TemplateSave 新增/编辑短信模板
		TemplateSave(ctx context.Context, in *adminin.SmsTemplateSaveInp) error
		// TemplateDelete 删除短信模板
		TemplateDelete(ctx context.Context, id uint64) error
		// TemplateTest 测试发送短信（解析变量并组装参数）
		TemplateTest(ctx context.Context, in *adminin.SmsTemplateTestInp) (*adminin.SmsTemplateTestModel, error)
		// VariableList 短信变量列表
		VariableList(ctx context.Context, in *adminin.SmsVariableListInp) (*adminin.SmsVariableListModel, error)
		// VariableSave 新增/编辑短信变量
		VariableSave(ctx context.Context, in *adminin.SmsVariableSaveInp) error
		// VariableDelete 删除短信变量
		VariableDelete(ctx context.Context, id uint64) error
		// LogList 发送日志列表
		LogList(ctx context.Context, in *adminin.SmsLogListInp) (*adminin.SmsLogListModel, error)
	}
)

var localSms ISms

func Sms() ISms {
	if localSms == nil {
		panic("implement not found for interface ISms, forgot register?")
	}
	return localSms
}

func RegisterSms(i ISms) {
	localSms = i
}
