package sms

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

// Aliyun 阿里云短信驱动
type Aliyun struct {
	conf   DriverConf
	client *dysmsapi.Client
}

func NewAliyun(conf DriverConf) (*Aliyun, error) {
	if conf.AccessId == "" || conf.AccessKey == "" {
		return nil, fmt.Errorf("aliyun sms: accessKeyId and accessKeySecret are required")
	}
	client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", conf.AccessId, conf.AccessKey)
	if err != nil {
		return nil, fmt.Errorf("aliyun sms: create client error: %v", err)
	}
	return &Aliyun{conf: conf, client: client}, nil
}

func (a *Aliyun) DriverName() string {
	return "aliyun"
}

func (a *Aliyun) Send(ctx context.Context, req *SendRequest) (*SendResult, error) {
	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = req.Phone
	request.TemplateCode = req.TemplateId

	signName := req.SignName
	if signName == "" {
		signName = a.conf.SignName
	}
	request.SignName = signName

	if len(req.Params) > 0 {
		paramJSON, err := json.Marshal(req.Params)
		if err != nil {
			return nil, fmt.Errorf("aliyun sms: marshal params error: %v", err)
		}
		request.TemplateParam = string(paramJSON)
	}

	resp, err := a.client.SendSms(request)
	if err != nil {
		return &SendResult{
			Success: false,
			Message: err.Error(),
		}, err
	}

	result := &SendResult{
		RequestId: resp.RequestId,
		Code:      resp.Code,
		Message:   resp.Message,
		Success:   resp.Code == "OK",
	}
	return result, nil
}
