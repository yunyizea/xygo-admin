package adminin

// ===================== 登录 =====================

// LoginInp 管理员登录入参
type LoginInp struct {
	Username    string `p:"username" v:"required#用户名不能为空" json:"username" dc:"用户名"`
	Password    string `p:"password" v:"required#密码不能为空" json:"password" dc:"密码"`
	CaptchaId   string `p:"captchaId" json:"captchaId" dc:"点选验证码ID"`
	CaptchaInfo string `p:"captchaInfo" json:"captchaInfo" dc:"点选验证码坐标信息"`
}

// LoginModel 管理员登录出参
type LoginModel struct {
	Id               uint64 `json:"id" dc:"用户ID"`
	Username         string `json:"username" dc:"用户名"`
	Nickname         string `json:"nickname" dc:"昵称"`
	AccessToken      string `json:"accessToken" dc:"访问令牌"`
	ExpiresIn        int64  `json:"expiresIn" dc:"访问令牌过期时间（秒）"`
	RefreshToken     string `json:"refreshToken" dc:"刷新令牌"`
	RefreshExpiresIn int64  `json:"refreshExpiresIn" dc:"刷新令牌过期时间（秒）"`
}

// ===================== 个人信息 =====================

// ProfileModel 个人信息响应模型
type ProfileModel struct {
	Id           uint64   `json:"id" dc:"用户ID"`
	Username     string   `json:"username" dc:"用户名"`
	Nickname     string   `json:"nickname" dc:"昵称"`
	RealName     string   `json:"realName" dc:"真实姓名"`
	Avatar       string   `json:"avatar" dc:"头像"`
	Email        string   `json:"email" dc:"邮箱"`
	Mobile       string   `json:"mobile" dc:"手机号"`
	Address      string   `json:"address" dc:"地址"`
	Remark       string   `json:"remark" dc:"个人简介"`
	Gender       int      `json:"gender" dc:"性别：0=未知 1=男 2=女"`
	DeptId       uint64   `json:"deptId" dc:"部门ID"`
	DeptName     string   `json:"deptName" dc:"部门名称"`
	DeptFullPath string   `json:"deptFullPath" dc:"部门全路径"`
	PostNames    []string `json:"postNames" dc:"岗位名称列表"`
	IsSuper      bool     `json:"isSuper" dc:"是否超级管理员"`
	Roles        []string `json:"roles" dc:"角色编码列表"`
	Buttons      []string `json:"buttons" dc:"按钮权限列表"`
}

// ===================== 更新个人信息 =====================

// UpdateProfileInp 更新个人信息入参
type UpdateProfileInp struct {
	Nickname string `json:"nickname" v:"required|length:2,50#请输入昵称|昵称长度为2-50个字符" dc:"昵称"`
	RealName string `json:"realName" dc:"真实姓名"`
	Avatar   string `json:"avatar" dc:"头像URL"`
	Email    string `json:"email" v:"email#邮箱格式不正确" dc:"邮箱"`
	Mobile   string `json:"mobile" v:"phone#手机号格式不正确" dc:"手机号"`
	Address  string `json:"address" dc:"地址"`
	Gender   int    `json:"gender" v:"in:0,1,2#性别值不正确" dc:"性别：0=未知 1=男 2=女"`
	Remark   string `json:"remark" v:"max-length:500#个人简介最多500字符" dc:"个人简介"`
}

// ===================== 修改密码 =====================

// ChangePasswordInp 修改密码入参
type ChangePasswordInp struct {
	OldPassword     string `json:"oldPassword" v:"required#请输入当前密码" dc:"当前密码"`
	NewPassword     string `json:"newPassword" v:"required|length:6,20#请输入新密码|密码长度为6-20个字符" dc:"新密码"`
	ConfirmPassword string `json:"confirmPassword" v:"required|same:NewPassword#请输入确认密码|两次密码输入不一致" dc:"确认密码"`
}
