package member

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"xygo/api/member"
	"xygo/internal/dao"
	"xygo/internal/library/token"
	"xygo/internal/model"
	"xygo/internal/model/input/memberin"
	"xygo/internal/service"
)

// Login 会员登录
func (c *ControllerV1) Login(ctx context.Context, req *member.LoginReq) (res *member.LoginRes, err error) {
	input := &memberin.LoginInput{
		Username:  req.Username,
		Password:  req.Password,
		Captcha:   req.Captcha,
		CaptchaId: req.CaptchaId,
	}

	output, err := service.MemberAuth().Login(ctx, input)
	if err != nil {
		return nil, err
	}

	return &member.LoginRes{
		Token:            output.Token,
		ExpiresIn:        output.ExpiresIn,
		RefreshToken:     output.RefreshToken,
		RefreshExpiresIn: output.RefreshExpiresIn,
	}, nil
}

// Refresh 刷新会员 accessToken
func (c *ControllerV1) Refresh(ctx context.Context, req *member.RefreshReq) (res *member.RefreshRes, err error) {
	userId, err := token.ValidateRefreshToken(ctx, token.AppMember, req.RefreshToken)
	if err != nil {
		return nil, gerror.New("刷新令牌无效或已过期，请重新登录")
	}

	var memberEntity *struct {
		Id       uint64  `json:"id"`
		Username string  `json:"username"`
		Nickname string  `json:"nickname"`
		Avatar   string  `json:"avatar"`
		Email    string  `json:"email"`
		Mobile   string  `json:"mobile"`
		Gender   int     `json:"gender"`
		Level    int     `json:"level"`
		GroupId  uint64  `json:"groupId"`
		Score    int     `json:"score"`
		Money    float64 `json:"money"`
		Status   int     `json:"status"`
	}
	if err = dao.Member.Ctx(ctx).Where("id", userId).Scan(&memberEntity); err != nil {
		return nil, err
	}
	if memberEntity == nil || memberEntity.Status != 1 {
		return nil, gerror.New("用户不存在或已禁用")
	}

	memberUser := model.MemberUser{
		Id:       memberEntity.Id,
		Username: memberEntity.Username,
		Nickname: memberEntity.Nickname,
		Avatar:   memberEntity.Avatar,
		Email:    memberEntity.Email,
		Mobile:   memberEntity.Mobile,
		Gender:   memberEntity.Gender,
		Level:    uint(memberEntity.Level),
		GroupId:  memberEntity.GroupId,
		Score:    memberEntity.Score,
		Money:    memberEntity.Money,
		LoginAt:  gtime.Now().Unix(),
	}

	accessToken, expiresIn, err := token.RefreshAccessMember(ctx, req.RefreshToken, memberUser)
	if err != nil {
		return nil, gerror.New("刷新令牌无效或已过期，请重新登录")
	}

	return &member.RefreshRes{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	}, nil
}

// Register 会员注册
func (c *ControllerV1) Register(ctx context.Context, req *member.RegisterReq) (res *member.RegisterRes, err error) {
	input := &memberin.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		Mobile:   req.Mobile,
		Email:    req.Email,
		Code:     req.Code,
	}

	output, err := service.MemberAuth().Register(ctx, input)
	if err != nil {
		return nil, err
	}

	return &member.RegisterRes{
		Id: output.Id,
	}, nil
}

// Logout 会员退出登录
func (c *ControllerV1) Logout(ctx context.Context, req *member.LogoutReq) (res *member.LogoutRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	if r != nil {
		tokenStr := r.Header.Get("Xy-User-Token")
		if tokenStr != "" {
			_ = service.MemberAuth().Logout(ctx, tokenStr)
		}
	}
	return &member.LogoutRes{}, nil
}

// 注：验证码接口已统一到公共 /captcha/click 和 /captcha/checkClick
