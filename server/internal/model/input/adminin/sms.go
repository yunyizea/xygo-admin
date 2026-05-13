package adminin

// ==================== 短信模板 ====================

type SmsTemplateListInp struct {
	Page   int
	Size   int
	Status int
	Code   string
	Title  string
}

type SmsTemplateListModel struct {
	List  []SmsTemplateListItem `json:"list"`
	Total int                   `json:"total"`
	Page  int                   `json:"page"`
	Size  int                   `json:"size"`
}

type SmsTemplateListItem struct {
	Id                 uint64      `json:"id"`
	Title              string      `json:"title"`
	Code               string      `json:"code"`
	Content            string      `json:"content"`
	ProviderTemplateId string      `json:"providerTemplateId"`
	Variables          interface{} `json:"variables"`
	RelatedVariableId  uint64      `json:"relatedVariableId"`
	Status             int         `json:"status"`
	Sort               int         `json:"sort"`
	Remark             string      `json:"remark"`
	CreateTime         uint64      `json:"createTime"`
	UpdateTime         uint64      `json:"updateTime"`
}

type SmsTemplateSaveInp struct {
	Id                 uint64
	Title              string
	Code               string
	Content            string
	ProviderTemplateId string
	Variables          string
	RelatedVariableId  uint64
	Status             int
	Sort               int
	Remark             string
}

type SmsTemplateTestInp struct {
	Id    uint64
	Phone string
}

type SmsTemplateTestModel struct {
	Success   bool   `json:"success"`
	RequestId string `json:"requestId"`
	Message   string `json:"message"`
}

// ==================== 短信变量 ====================

type SmsVariableListInp struct {
	Page int
	Size int
	Name string
}

type SmsVariableListModel struct {
	List  []SmsVariableListItem `json:"list"`
	Total int                   `json:"total"`
	Page  int                   `json:"page"`
	Size  int                   `json:"size"`
}

type SmsVariableListItem struct {
	Id          uint64 `json:"id"`
	Title       string `json:"title"`
	Name        string `json:"name"`
	SourceType  int    `json:"sourceType"`
	SqlQuery    string `json:"sqlQuery"`
	MethodName  string `json:"methodName"`
	SharedCount int    `json:"sharedCount"`
	Status      int    `json:"status"`
	CreateTime  uint64 `json:"createTime"`
	UpdateTime  uint64 `json:"updateTime"`
}

type SmsVariableSaveInp struct {
	Id         uint64
	Title      string
	Name       string
	SourceType int
	SqlQuery   string
	MethodName string
	Status     int
}

// ==================== 短信日志 ====================

type SmsLogListInp struct {
	Page         int
	Size         int
	Phone        string
	TemplateCode string
	Status       int
	Driver       string
}

type SmsLogListModel struct {
	List  []SmsLogListItem `json:"list"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
}

type SmsLogListItem struct {
	Id           uint64      `json:"id"`
	Phone        string      `json:"phone"`
	TemplateCode string      `json:"templateCode"`
	Driver       string      `json:"driver"`
	Content      string      `json:"content"`
	Params       interface{} `json:"params"`
	Status       int         `json:"status"`
	RequestId    string      `json:"requestId"`
	ErrorMsg     string      `json:"errorMsg"`
	CreateTime   uint64      `json:"createTime"`
}
