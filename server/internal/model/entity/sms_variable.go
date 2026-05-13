// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// SmsVariable is the golang structure for table sms_variable.
type SmsVariable struct {
	Id          uint64 `json:"id"          orm:"id"           description:"主键"`
	Title       string `json:"title"       orm:"title"        description:"变量标题"`
	Name        string `json:"name"        orm:"name"         description:"变量名"`
	SourceType  int    `json:"sourceType"  orm:"source_type"  description:"来源类型：1=字段提取 2=SQL查询 3=内置Helper"`
	SqlQuery    string `json:"sqlQuery"    orm:"sql_query"    description:"SQL查询语句"`
	MethodName  string `json:"methodName"  orm:"method_name"  description:"Helper方法路径"`
	SharedCount int    `json:"sharedCount" orm:"shared_count" description:"共通数据数"`
	Status      int    `json:"status"      orm:"status"       description:"状态：1=启用 0=禁用"`
	CreatedBy   uint64 `json:"createdBy"   orm:"created_by"   description:"创建人ID"`
	UpdatedBy   uint64 `json:"updatedBy"   orm:"updated_by"   description:"更新人ID"`
	CreateTime  uint64 `json:"createTime"  orm:"create_time"  description:"创建时间"`
	UpdateTime  uint64 `json:"updateTime"  orm:"update_time"  description:"更新时间"`
}
