package admin

import (
	"github.com/gogf/gf/v2/frame/g"
)

// FieldPermListReq 查询字段权限列表
type FieldPermListReq struct {
	g.Meta   `path:"/admin/fieldPerm/list" method:"get" tags:"FieldPerm" summary:"字段权限列表"`
	RoleId   uint64 `json:"roleId" dc:"角色ID"`
	Module   string `json:"module" dc:"模块"`
	Resource string `json:"resource" dc:"资源标识"`
}

type FieldPermListRes struct {
	List []FieldPermItem `json:"list"`
}

type FieldPermItem struct {
	Id         uint64 `json:"id"`
	RoleId     uint64 `json:"roleId"`
	Module     string `json:"module"`
	Resource   string `json:"resource"`
	FieldName  string `json:"fieldName"`
	FieldLabel string `json:"fieldLabel"`
	PermType   int    `json:"permType"`
	Status     int    `json:"status"`
	Remark     string `json:"remark"`
}

// FieldPermBatchSaveReq 批量保存字段权限
type FieldPermBatchSaveReq struct {
	g.Meta   `path:"/admin/fieldPerm/batchSave" method:"post" tags:"FieldPerm" summary:"批量保存字段权限"`
	RoleId   uint64            `json:"roleId" v:"required#角色ID不能为空"`
	Resource string            `json:"resource" v:"required#资源标识不能为空"`
	Fields   []FieldPermConfig `json:"fields" v:"required#字段配置不能为空"`
}

type FieldPermConfig struct {
	FieldName  string `json:"fieldName" v:"required#字段名称不能为空"`
	FieldLabel string `json:"fieldLabel"`
	PermType   int    `json:"permType" v:"required|between:0,2#权限类型不能为空|权限类型值错误"`
}

type FieldPermBatchSaveRes struct{}

// FieldPermGetByRoleReq 获取角色的字段权限映射
type FieldPermGetByRoleReq struct {
	g.Meta   `path:"/admin/fieldPerm/getByRole" method:"get" tags:"FieldPerm" summary:"获取角色字段权限"`
	RoleId   uint64 `json:"roleId" v:"required#角色ID不能为空"`
	Resource string `json:"resource" dc:"资源标识（可选）"`
}

type FieldPermGetByRoleRes struct {
	FieldPerms map[string]map[string]int `json:"fieldPerms"` // resource -> field -> permType
}

// GetResourceFieldsReq 获取资源的字段列表
type GetResourceFieldsReq struct {
	g.Meta   `path:"/admin/fieldPerm/resourceFields" method:"get" tags:"FieldPerm" summary:"获取资源字段列表"`
	Resource string `p:"resource" json:"resource" v:"required#资源标识不能为空"`
}

type GetResourceFieldsRes struct {
	Fields []ResourceFieldItem `json:"fields"`
}

type ResourceFieldItem struct {
	FieldName   string `json:"fieldName"`
	FieldLabel  string `json:"fieldLabel"`
	IsSensitive bool   `json:"isSensitive"`
}

// FieldPermMineReq 获取当前登录用户的字段权限（合并所有角色，取最高权限）
type FieldPermMineReq struct {
	g.Meta `path:"/admin/fieldPerm/mine" method:"get" tags:"FieldPerm" summary:"获取当前用户字段权限"`
}

type FieldPermMineRes struct {
	FieldPerms map[string]map[string]int `json:"fieldPerms"` // resource -> field -> permType
}
