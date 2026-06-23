package gencli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// flags 命令行解析结果。
type flags struct {
	table      string
	genType    int
	varName    string
	head       string
	column     string
	auto       string
	menuPid    int
	icon       string
	sort       int
	view       string
	treePid    string
	treeTitle  string
	spec       string
	preview    bool
	noGf       bool
	listTables bool
	yes        bool
}

// optionsJSON 生成器 options，键名与 logic/gencodes.parseOptions 对齐。
type optionsJSON struct {
	GenType   int               `json:"genType"`
	HeadOps   []string          `json:"headOps"`
	ColumnOps []string          `json:"columnOps"`
	AutoOps   []string          `json:"autoOps"`
	ApiPrefix string            `json:"apiPrefix,omitempty"`
	GenPaths  map[string]string `json:"genPaths,omitempty"`
	AddonName string            `json:"addonName,omitempty"`
	Menu      struct {
		Pid  int    `json:"pid"`
		Icon string `json:"icon"`
		Sort int    `json:"sort"`
	} `json:"menu"`
	ViewMode string `json:"viewMode"`
	Tree     struct {
		TitleColumn string `json:"titleColumn"`
		PidColumn   string `json:"pidColumn"`
	} `json:"tree"`
}

// specFile 精确模式 JSON 结构。
type specFile struct {
	Table   string                `json:"table"`
	Var     string                `json:"var"`
	GenType int                   `json:"genType"`
	Options *specOptions          `json:"options"`
	Columns map[string]specColumn `json:"columns"`
}

type specOptions struct {
	HeadOps   []string          `json:"headOps"`
	ColumnOps []string          `json:"columnOps"`
	AutoOps   []string          `json:"autoOps"`
	ApiPrefix string            `json:"apiPrefix"`
	AddonName string            `json:"addonName"`
	GenPaths  map[string]string `json:"genPaths"`
	Menu      struct {
		Pid  int    `json:"pid"`
		Icon string `json:"icon"`
		Sort int    `json:"sort"`
	} `json:"menu"`
	ViewMode string `json:"viewMode"`
	Tree     struct {
		PidColumn   string `json:"pidColumn"`
		TitleColumn string `json:"titleColumn"`
	} `json:"tree"`
}

// specColumn 字段增量覆盖；指针表示"仅当提供时覆盖"。
type specColumn struct {
	DesignType *string         `json:"designType"`
	FormType   *string         `json:"formType"`
	QueryType  *string         `json:"queryType"`
	IsList     *int            `json:"isList"`
	IsEdit     *int            `json:"isEdit"`
	IsQuery    *int            `json:"isQuery"`
	IsRequired *int            `json:"isRequired"`
	DictType   *string         `json:"dictType"`
	Extra      json.RawMessage `json:"extra"`
}

// 默认操作项（与 web 设计器默认勾选一致）。
var (
	defaultHeadOps   = []string{"add", "batchDel", "export"}
	defaultColumnOps = []string{"edit", "del", "view", "status", "check"}
	defaultAutoOps   = []string{"genMenuPermissions", "runDao", "runService"}
)

// defaultMenuIcon 默认菜单图标。
// 菜单图标由前端 ArtSvgIcon(@iconify/vue) 渲染，只认 Iconify 名称，
// 项目约定使用 Remix 图标集（ri: 前缀），如 ri:box-3-line。
// 不能用 ele-Document 这类 Element Plus 命名，否则显示空白。
const defaultMenuIcon = "ri:apps-2-line"

// parseFlags 手动解析 gen 之后的参数，支持 --k v / --k=v / --flag / 位置参数(表名)。
func parseFlags(args []string) (*flags, error) {
	f := &flags{genType: 10}

	// 需要取值的字符串/数字 flag
	next := func(i int, key string) (string, int, error) {
		if i+1 >= len(args) {
			return "", i, fmt.Errorf("flag %s 缺少值", key)
		}
		return args[i+1], i + 1, nil
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// 位置参数：第一个非 -- 开头的当作表名
		if !strings.HasPrefix(arg, "-") {
			if f.table == "" {
				f.table = arg
			}
			continue
		}

		key := arg
		var inlineVal string
		var hasInline bool
		if idx := strings.Index(arg, "="); idx >= 0 {
			key = arg[:idx]
			inlineVal = arg[idx+1:]
			hasInline = true
		}

		getVal := func() (string, error) {
			if hasInline {
				return inlineVal, nil
			}
			v, ni, err := next(i, key)
			if err != nil {
				return "", err
			}
			i = ni
			return v, nil
		}

		var err error
		switch key {
		case "--preview":
			f.preview = true
		case "--no-gf":
			f.noGf = true
		case "--list-tables":
			f.listTables = true
		case "--yes", "-y":
			f.yes = true
		case "--table":
			f.table, err = getVal()
		case "--type":
			var v string
			if v, err = getVal(); err == nil {
				f.genType, _ = strconv.Atoi(v)
			}
		case "--var":
			f.varName, err = getVal()
		case "--head":
			f.head, err = getVal()
		case "--column":
			f.column, err = getVal()
		case "--auto":
			f.auto, err = getVal()
		case "--menu-pid":
			var v string
			if v, err = getVal(); err == nil {
				f.menuPid, _ = strconv.Atoi(v)
			}
		case "--icon":
			f.icon, err = getVal()
		case "--sort":
			var v string
			if v, err = getVal(); err == nil {
				f.sort, _ = strconv.Atoi(v)
			}
		case "--view":
			f.view, err = getVal()
		case "--tree-pid":
			f.treePid, err = getVal()
		case "--tree-title":
			f.treeTitle, err = getVal()
		case "--spec":
			f.spec, err = getVal()
		default:
			return nil, fmt.Errorf("未知 flag: %s", key)
		}
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}
