package sms

import (
	"context"
	"fmt"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

// Tencent 腾讯云短信驱动
type Tencent struct {
	conf   DriverConf
	client *tencentSms.Client
}

func NewTencent(conf DriverConf) (*Tencent, error) {
	if conf.AccessId == "" || conf.AccessKey == "" {
		return nil, fmt.Errorf("tencent sms: secretId and secretKey are required")
	}
	credential := common.NewCredential(conf.AccessId, conf.AccessKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"

	client, err := tencentSms.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		return nil, fmt.Errorf("tencent sms: create client error: %v", err)
	}
	return &Tencent{conf: conf, client: client}, nil
}

func (t *Tencent) DriverName() string {
	return "tencent"
}

func (t *Tencent) Send(ctx context.Context, req *SendRequest) (*SendResult, error) {
	request := tencentSms.NewSendSmsRequest()

	appId := t.conf.Extra["appId"]
	if appId == "" {
		return nil, fmt.Errorf("tencent sms: appId is required")
	}
	request.SmsSdkAppId = common.StringPtr(appId)

	signName := req.SignName
	if signName == "" {
		signName = t.conf.SignName
	}
	request.SignName = common.StringPtr(signName)
	request.TemplateId = common.StringPtr(req.TemplateId)

	phone := req.Phone
	if !strings.HasPrefix(phone, "+") {
		phone = "+86" + phone
	}
	request.PhoneNumberSet = common.StringPtrs([]string{phone})

	if len(req.ParamList) > 0 {
		request.TemplateParamSet = common.StringPtrs(req.ParamList)
	} else if len(req.Params) > 0 {
		values := make([]string, 0, len(req.Params))
		for i := 1; i <= len(req.Params); i++ {
			key := fmt.Sprintf("%d", i)
			if v, ok := req.Params[key]; ok {
				values = append(values, v)
			}
		}
		if len(values) == 0 {
			for _, v := range req.Params {
				values = append(values, v)
			}
		}
		request.TemplateParamSet = common.StringPtrs(values)
	}

	resp, err := t.client.SendSms(request)
	if err != nil {
		return &SendResult{
			Success: false,
			Message: err.Error(),
		}, err
	}

	result := &SendResult{
		RequestId: *resp.Response.RequestId,
	}

	if len(resp.Response.SendStatusSet) > 0 {
		status := resp.Response.SendStatusSet[0]
		result.Code = *status.Code
		result.Message = *status.Message
		result.Success = *status.Code == "Ok"
	}

	return result, nil
}
