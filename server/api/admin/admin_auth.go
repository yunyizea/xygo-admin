package admin

import (
	"github.com/gogf/gf/v2/frame/g"

	"xygo/internal/model/input/adminin"
)

// ===================== 登录 =====================

type LoginReq struct {
	g.Meta `path:"/admin/auth/login" method:"post" tags:"AdminAuth" summary:"Admin login"`
	adminin.LoginInp
}

type LoginRes struct {
	*adminin.LoginModel
}

// ===================== 刷新令牌 =====================

type RefreshReq struct {
	g.Meta       `path:"/admin/auth/refresh" method:"post" tags:"AdminAuth" summary:"Refresh access token"`
	RefreshToken string `json:"refreshToken" v:"required#刷新令牌不能为空"`
}

type RefreshRes struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}

// ===================== 登出 =====================

type LogoutReq struct {
	g.Meta `path:"/admin/auth/logout" method:"post" tags:"AdminAuth" summary:"Admin logout"`
}

type LogoutRes struct{}

// ===================== 个人信息 =====================

type ProfileReq struct {
	g.Meta `path:"/admin/auth/profile" method:"get" tags:"AdminAuth" summary:"Get current admin profile"`
}

type ProfileRes struct {
	*adminin.ProfileModel
}

// ===================== 更新个人信息 =====================

type UpdateProfileReq struct {
	g.Meta `path:"/admin/auth/updateProfile" method:"post" tags:"AdminAuth" summary:"Update current user profile"`
	adminin.UpdateProfileInp
}

type UpdateProfileRes struct{}

// ===================== 修改密码 =====================

type ChangePasswordReq struct {
	g.Meta `path:"/admin/auth/changePassword" method:"post" tags:"AdminAuth" summary:"Change current user password"`
	adminin.ChangePasswordInp
}

type ChangePasswordRes struct{}
