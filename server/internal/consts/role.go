package consts

// ============================================
// 数据范围常量
// ============================================

// 数据范围类型（DataScope）
const (
	RoleDataAll = 1 // 全部权限

	// 通过部门划分
	RoleDataNowDept    = 2 // 当前部门
	RoleDataDeptAndSub = 3 // 当前部门及以下部门
	RoleDataDeptCustom = 4 // 自定义部门

	// 通过上下级关系划分
	RoleDataSelf          = 5 // 仅自己
	RoleDataSelfAndSub    = 6 // 自己和直属下级
	RoleDataSelfAndAllSub = 7 // 自己和全部下级
)

// RoleDataNameMap 数据范围名称映射
var RoleDataNameMap = map[int]string{
	RoleDataAll:           "全部权限",
	RoleDataNowDept:       "当前部门",
	RoleDataDeptAndSub:    "当前及以下部门",
	RoleDataDeptCustom:    "自定义部门",
	RoleDataSelf:          "仅自己",
	RoleDataSelfAndSub:    "自己和直属下级",
	RoleDataSelfAndAllSub: "自己和全部下级",
}

// ============================================
// 超级管理员常量
// ============================================

// SuperRoleKey 超级管理员角色标识（唯一字符串）
// 判断超级管理员时只用此常量（role.key），不使用用户级别的 is_super 字段
const SuperRoleKey = "super_admin"

// SuperRoleKeys 所有超级管理员角色标识（如有多个超管角色）
var SuperRoleKeys = []string{
	"super_admin",
	// 如果有其他超管角色，在此添加
	// "root",
	// "super",
}

// IsSuperRole 判断角色标识是否为超级管理员
// 只基于角色 key 判断，不使用用户的 is_super 字段
func IsSuperRole(roleKey string) bool {
	for _, key := range SuperRoleKeys {
		if roleKey == key {
			return true
		}
	}
	return false
}

// ============================================
// 数据范围选项（用于前端下拉选择）
// ============================================

// DataScopeOption 数据范围选项
type DataScopeOption struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}

// GroupDataScopeOption 分组的数据范围选项
type GroupDataScopeOption struct {
	Type     string            `json:"type,omitempty"`     // "group" 表示分组
	Label    string            `json:"label"`              // 分组或选项标签
	Key      int               `json:"key"`                // 分组 key（负数）
	Value    int               `json:"value,omitempty"`    // 选项值
	Children []DataScopeOption `json:"children,omitempty"` // 子选项
}

// DataScopeSelect 数据范围选择列表（用于前端下拉框）
var DataScopeSelect = []GroupDataScopeOption{
	{
		Label: RoleDataNameMap[RoleDataAll],
		Key:   RoleDataAll,
		Value: RoleDataAll,
	},
	{
		Type:  "group",
		Label: "按部门划分",
		Key:   -1,
		Children: []DataScopeOption{
			{
				Label: RoleDataNameMap[RoleDataNowDept],
				Value: RoleDataNowDept,
			},
			{
				Label: RoleDataNameMap[RoleDataDeptAndSub],
				Value: RoleDataDeptAndSub,
			},
			{
				Label: RoleDataNameMap[RoleDataDeptCustom],
				Value: RoleDataDeptCustom,
			},
		},
	},
	{
		Type:  "group",
		Label: "按上下级关系划分",
		Key:   -2,
		Children: []DataScopeOption{
			{
				Label: RoleDataNameMap[RoleDataSelf],
				Value: RoleDataSelf,
			},
			{
				Label: RoleDataNameMap[RoleDataSelfAndSub],
				Value: RoleDataSelfAndSub,
			},
			{
				Label: RoleDataNameMap[RoleDataSelfAndAllSub],
				Value: RoleDataSelfAndAllSub,
			},
		},
	},
}

// ============================================
// 部门类型常量（当前项目只使用普通部门，不需要多租户）
// ============================================

// 注意：如需多租户支持，安装 tenant 插件即可，
// 无需修改核心代码
