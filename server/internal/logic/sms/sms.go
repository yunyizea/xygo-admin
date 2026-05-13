package sms

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"xygo/internal/consts"
	"xygo/internal/dao"
	smsDrv "xygo/internal/library/sms"
	"xygo/internal/model/do"
	"xygo/internal/model/entity"
	"xygo/internal/model/input/adminin"
	"xygo/internal/service"
	"xygo/utility"
)

type sSms struct{}

func init() {
	service.RegisterSms(&sSms{})
}

// ==================== 模板 ====================

func (s *sSms) TemplateList(ctx context.Context, in *adminin.SmsTemplateListInp) (*adminin.SmsTemplateListModel, error) {
	m := dao.SmsTemplate.Ctx(ctx)

	if in.Status >= 0 {
		m = m.Where("status", in.Status)
	}
	if in.Code != "" {
		m = m.WhereLike("code", "%"+in.Code+"%")
	}
	if in.Title != "" {
		m = m.WhereLike("title", "%"+in.Title+"%")
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var items []entity.SmsTemplate
	if err = m.OrderAsc("sort").OrderAsc("id").Page(in.Page, in.Size).Scan(&items); err != nil {
		return nil, err
	}

	list := make([]adminin.SmsTemplateListItem, 0, len(items))
	for _, it := range items {
		list = append(list, adminin.SmsTemplateListItem{
			Id:                 it.Id,
			Title:              it.Title,
			Code:               it.Code,
			Content:            it.Content,
			ProviderTemplateId: it.ProviderTemplateId,
			Variables:          it.Variables,
			RelatedVariableId:  it.RelatedVariableId,
			Status:             it.Status,
			Sort:               it.Sort,
			Remark:             it.Remark,
			CreateTime:         it.CreateTime,
			UpdateTime:         it.UpdateTime,
		})
	}

	return &adminin.SmsTemplateListModel{List: list, Total: total, Page: in.Page, Size: in.Size}, nil
}

func (s *sSms) TemplateSave(ctx context.Context, in *adminin.SmsTemplateSaveInp) error {
	now := utility.NowUnix()

	var variables *gjson.Json
	if in.Variables != "" {
		var err error
		variables, err = gjson.LoadContent([]byte(in.Variables))
		if err != nil {
			return gerror.NewCode(consts.CodeInvalidParam, "variables 非法 JSON")
		}
	}

	if in.Id > 0 {
		count, err := dao.SmsTemplate.Ctx(ctx).Where("code", in.Code).WhereNot("id", in.Id).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return gerror.NewCode(consts.CodeDuplicateData, fmt.Sprintf("模板标识已存在：%s", in.Code))
		}
		_, err = dao.SmsTemplate.Ctx(ctx).Data(do.SmsTemplate{
			Title: in.Title, Code: in.Code, Content: in.Content,
			ProviderTemplateId: in.ProviderTemplateId, Variables: variables,
			RelatedVariableId: in.RelatedVariableId, Status: in.Status,
			Sort: in.Sort, Remark: in.Remark, UpdateTime: now,
		}).Where("id", in.Id).Update()
		return err
	}

	count, err := dao.SmsTemplate.Ctx(ctx).Where("code", in.Code).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.NewCode(consts.CodeDuplicateData, fmt.Sprintf("模板标识已存在：%s", in.Code))
	}
	_, err = dao.SmsTemplate.Ctx(ctx).Data(do.SmsTemplate{
		Title: in.Title, Code: in.Code, Content: in.Content,
		ProviderTemplateId: in.ProviderTemplateId, Variables: variables,
		RelatedVariableId: in.RelatedVariableId, Status: in.Status,
		Sort: in.Sort, Remark: in.Remark, CreateTime: now, UpdateTime: now,
	}).Insert()
	return err
}

func (s *sSms) TemplateDelete(ctx context.Context, id uint64) error {
	_, err := dao.SmsTemplate.Ctx(ctx).Where("id", id).Delete()
	return err
}

// TemplateTest 测试发送：解析模板中的 ${var} 占位符，按序号填充测试值
func (s *sSms) TemplateTest(ctx context.Context, in *adminin.SmsTemplateTestInp) (*adminin.SmsTemplateTestModel, error) {
	var tpl entity.SmsTemplate
	if err := dao.SmsTemplate.Ctx(ctx).Where("id", in.Id).Scan(&tpl); err != nil {
		return nil, err
	}
	if tpl.Id == 0 {
		return nil, gerror.NewCode(consts.CodeDataNotFound, "模板不存在")
	}

	params, paramList := s.buildTestParams(tpl)

	mgr := smsDrv.Instance(ctx)
	result, err := mgr.Send(ctx, &smsDrv.SendRequest{
		Phone:      in.Phone,
		TemplateId: tpl.ProviderTemplateId,
		Params:     params,
		ParamList:  paramList,
	})
	if err != nil {
		return &adminin.SmsTemplateTestModel{Success: false, Message: err.Error()}, nil
	}

	s.saveLog(ctx, in.Phone, tpl.Code, tpl.Content, result)

	return &adminin.SmsTemplateTestModel{
		Success:   result.Success,
		RequestId: result.RequestId,
		Message:   result.Message,
	}, nil
}

// buildTestParams 从模板 content 中提取 ${var} 占位符，构造测试参数
// 阿里云用 key-value map（key=变量名, value="test"），腾讯云用有序列表
var varRegex = regexp.MustCompile(`\$\{(\w+)\}`)

func (s *sSms) buildTestParams(tpl entity.SmsTemplate) (map[string]string, []string) {
	params := make(map[string]string)
	var paramList []string

	// 从 variables JSON 数组获取变量名
	varNames := make([]string, 0)
	if tpl.Variables != nil {
		arr := tpl.Variables.Array()
		for _, v := range arr {
			if name, ok := v.(string); ok && name != "" {
				varNames = append(varNames, name)
			}
		}
	}

	// 如果 variables 为空，从 content 中提取 ${xxx}
	if len(varNames) == 0 {
		matches := varRegex.FindAllStringSubmatch(tpl.Content, -1)
		for _, m := range matches {
			varNames = append(varNames, m[1])
		}
	}

	for i, name := range varNames {
		testVal := fmt.Sprintf("test%d", i+1)
		params[name] = testVal
		paramList = append(paramList, testVal)
	}

	// 如果完全没有变量，也给一个空 map（避免 nil）
	if len(params) == 0 {
		params = nil
	}

	return params, paramList
}

func (s *sSms) saveLog(ctx context.Context, phone, code, content string, result *smsDrv.SendResult) {
	status := 0
	if result.Success {
		status = 1
	}
	errMsg := ""
	if !result.Success {
		errMsg = result.Message
	}
	_, err := dao.SmsLog.Ctx(ctx).Data(do.SmsLog{
		Phone: phone, TemplateCode: code, Driver: result.Driver,
		Content: content, Status: status, RequestId: result.RequestId,
		ErrorMsg: errMsg, CreateTime: utility.NowUnix(),
	}).Insert()
	if err != nil {
		g.Log().Warningf(ctx, "[SMS] save log error: %v", err)
	}
}

// ==================== 变量 ====================

func (s *sSms) VariableList(ctx context.Context, in *adminin.SmsVariableListInp) (*adminin.SmsVariableListModel, error) {
	m := dao.SmsVariable.Ctx(ctx)
	if in.Name != "" {
		m = m.WhereLike("name", "%"+in.Name+"%")
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var items []entity.SmsVariable
	if err = m.OrderAsc("id").Page(in.Page, in.Size).Scan(&items); err != nil {
		return nil, err
	}

	list := make([]adminin.SmsVariableListItem, 0, len(items))
	for _, it := range items {
		list = append(list, adminin.SmsVariableListItem{
			Id: it.Id, Title: it.Title, Name: it.Name,
			SourceType: it.SourceType, SqlQuery: it.SqlQuery,
			MethodName: it.MethodName, SharedCount: it.SharedCount,
			Status: it.Status, CreateTime: it.CreateTime, UpdateTime: it.UpdateTime,
		})
	}

	return &adminin.SmsVariableListModel{List: list, Total: total, Page: in.Page, Size: in.Size}, nil
}

func (s *sSms) VariableSave(ctx context.Context, in *adminin.SmsVariableSaveInp) error {
	now := utility.NowUnix()

	if in.Id > 0 {
		count, err := dao.SmsVariable.Ctx(ctx).Where("name", in.Name).WhereNot("id", in.Id).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return gerror.NewCode(consts.CodeDuplicateData, fmt.Sprintf("变量名已存在：%s", in.Name))
		}
		_, err = dao.SmsVariable.Ctx(ctx).Data(do.SmsVariable{
			Title: in.Title, Name: in.Name, SourceType: in.SourceType,
			SqlQuery: in.SqlQuery, MethodName: in.MethodName,
			Status: in.Status, UpdateTime: now,
		}).Where("id", in.Id).Update()
		return err
	}

	count, err := dao.SmsVariable.Ctx(ctx).Where("name", in.Name).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.NewCode(consts.CodeDuplicateData, fmt.Sprintf("变量名已存在：%s", in.Name))
	}
	_, err = dao.SmsVariable.Ctx(ctx).Data(do.SmsVariable{
		Title: in.Title, Name: in.Name, SourceType: in.SourceType,
		SqlQuery: in.SqlQuery, MethodName: in.MethodName,
		Status: in.Status, CreateTime: now, UpdateTime: now,
	}).Insert()
	return err
}

func (s *sSms) VariableDelete(ctx context.Context, id uint64) error {
	_, err := dao.SmsVariable.Ctx(ctx).Where("id", id).Delete()
	return err
}

// ==================== 日志 ====================

func (s *sSms) LogList(ctx context.Context, in *adminin.SmsLogListInp) (*adminin.SmsLogListModel, error) {
	m := dao.SmsLog.Ctx(ctx)
	if in.Phone != "" {
		m = m.WhereLike("phone", "%"+in.Phone+"%")
	}
	if in.TemplateCode != "" {
		m = m.Where("template_code", in.TemplateCode)
	}
	if in.Status >= 0 {
		m = m.Where("status", in.Status)
	}
	if in.Driver != "" {
		m = m.Where("driver", in.Driver)
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var items []entity.SmsLog
	if err = m.OrderDesc("id").Page(in.Page, in.Size).Scan(&items); err != nil {
		return nil, err
	}

	list := make([]adminin.SmsLogListItem, 0, len(items))
	for _, it := range items {
		list = append(list, adminin.SmsLogListItem{
			Id: it.Id, Phone: it.Phone, TemplateCode: it.TemplateCode,
			Driver: it.Driver, Content: it.Content, Params: it.Params,
			Status: it.Status, RequestId: it.RequestId, ErrorMsg: it.ErrorMsg,
			CreateTime: it.CreateTime,
		})
	}

	return &adminin.SmsLogListModel{List: list, Total: total, Page: in.Page, Size: in.Size}, nil
}

// ==================== 配置钩子 ====================

// OnConfigChanged 当 sms 分组配置保存后调用，重置驱动单例
func OnConfigChanged(group string) {
	if strings.EqualFold(group, "sms") {
		smsDrv.ResetInstance()
	}
}
