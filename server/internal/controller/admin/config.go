// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package admin

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"

	api "xygo/api/admin"
	"xygo/internal/consts"
	"xygo/internal/dao"
	smsLogic "xygo/internal/logic/sms"
	"xygo/internal/model/do"
	"xygo/internal/model/entity"
	"xygo/internal/model/input/adminin"
	"xygo/utility"
)

// ConfigSchema 返回配置项 schema（含当前值）
func (c *ControllerV1) ConfigSchema(ctx context.Context, req *api.ConfigSchemaReq) (res *api.ConfigSchemaRes, err error) {
	var items []entity.SysConfig
	if err = dao.SysConfig.Ctx(ctx).
		OrderAsc(dao.SysConfig.Columns().Group).OrderAsc("sort").OrderAsc("id").
		Scan(&items); err != nil {
		return nil, err
	}

	result := make([]adminin.ConfigSchemaItem, 0, len(items))
	for _, it := range items {
		result = append(result, adminin.ConfigSchemaItem{
			Id:        it.Id,
			Group:     it.Group,
			GroupName: it.GroupName,
			Name:      it.Name,
			Key:       it.Key,
			Value:     it.Value,
			Type:      it.Type,
			Options:   it.Options,
			Rules:     it.Rules,
			Sort:      it.Sort,
			Remark:    it.Remark,
			AllowDel:  it.AllowDel,
		})
	}
	res = &api.ConfigSchemaRes{List: result}
	return
}

// ConfigList 获取指定分组配置
func (c *ControllerV1) ConfigList(ctx context.Context, req *api.ConfigListReq) (res *api.ConfigListRes, err error) {
	if req.Group == "" {
		return nil, gerror.NewCode(consts.CodeInvalidParam, "分组必填")
	}

	itemsMap, items, err := getConfigGroup(ctx, req.Group)
	if err != nil {
		return nil, err
	}

	kvList := make([]adminin.ConfigKVItem, 0, len(itemsMap))
	for k, v := range itemsMap {
		kvList = append(kvList, adminin.ConfigKVItem{Key: k, Value: v})
	}

	// 保持与 schema 顺序一致
	if len(items) > 0 {
		ordered := make([]adminin.ConfigKVItem, 0, len(items))
		for _, it := range items {
			ordered = append(ordered, adminin.ConfigKVItem{Key: it.Key, Value: itemsMap[it.Key]})
		}
		kvList = ordered
	}

	res = &api.ConfigListRes{
		ConfigListModel: adminin.ConfigListModel{
			Group: req.Group,
			Items: itemsMap,
			List:  kvList,
		},
	}
	return
}

// ConfigSave 保存分组配置
func (c *ControllerV1) ConfigSave(ctx context.Context, req *api.ConfigSaveReq) (res *api.ConfigSaveRes, err error) {
	if req.Group == "" {
		return nil, gerror.NewCode(consts.CodeInvalidParam, "分组必填")
	}
	if len(req.Items) == 0 {
		return nil, gerror.NewCode(consts.CodeInvalidParam, "配置列表不能为空")
	}

	// 拉取当前分组配置定义
	var items []entity.SysConfig
	if err = dao.SysConfig.Ctx(ctx).
		Where(dao.SysConfig.Columns().Group, req.Group).
		OrderAsc(dao.SysConfig.Columns().Group).OrderAsc("sort").OrderAsc("id").
		Scan(&items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, gerror.NewCode(consts.CodeDataNotFound, "分组不存在或无配置项")
	}

	itemMap := make(map[string]entity.SysConfig, len(items))
	for _, it := range items {
		itemMap[it.Key] = it
	}

	// 校验并准备写入数据
	type kv struct {
		key string
		val string
	}
	writeList := make([]kv, 0, len(req.Items))

	for _, in := range req.Items {
		item, ok := itemMap[in.Key]
		if !ok {
			return nil, gerror.NewCode(consts.CodeInvalidParam, fmt.Sprintf("配置键不存在：%s", in.Key))
		}
		normVal, err := normalizeValueByType(item.Type, in.Value)
		if err != nil {
			return nil, gerror.NewCode(consts.CodeInvalidParam, fmt.Sprintf("%s: %v", item.Name, err))
		}
		writeList = append(writeList, kv{key: in.Key, val: normVal})
	}

	// 事务写入
	err = dao.SysConfig.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, w := range writeList {
			_, err := dao.SysConfig.Ctx(ctx).
				Data(do.SysConfig{
					Value:      w.val,
					UpdateTime: utility.NowUnix(),
				}).
				Where(dao.SysConfig.Columns().Group, req.Group).
				Where(dao.SysConfig.Columns().Key, w.key).
				Update()
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 刷新缓存
	// 无缓存模式，保存后无需清理，后续读取直接查库

	// 通知各模块配置变更（如 sms 分组变更后重置驱动单例）
	smsLogic.OnConfigChanged(req.Group)

	res = &api.ConfigSaveRes{}
	return
}

// ConfigCreate 创建配置项
func (c *ControllerV1) ConfigCreate(ctx context.Context, req *api.ConfigCreateReq) (res *api.ConfigCreateRes, err error) {
	// 检查键唯一
	if count, err := dao.SysConfig.Ctx(ctx).Where(dao.SysConfig.Columns().Key, req.Key).Count(); err != nil {
		return nil, err
	} else if count > 0 {
		return nil, gerror.NewCode(consts.CodeInvalidParam, fmt.Sprintf("配置键已存在：%s", req.Key))
	}

	// options/rules 需要是合法 JSON 字符串
	var optJSON, rulesJSON *gjson.Json
	if req.Options != "" {
		j, err := gjson.LoadContent([]byte(req.Options))
		if err != nil {
			return nil, gerror.NewCode(consts.CodeInvalidParam, fmt.Sprintf("options 非法 JSON: %v", err))
		}
		optJSON = j
	}
	if req.Rules != "" {
		j, err := gjson.LoadContent([]byte(req.Rules))
		if err != nil {
			return nil, gerror.NewCode(consts.CodeInvalidParam, fmt.Sprintf("rules 非法 JSON: %v", err))
		}
		rulesJSON = j
	}

	// array / array-json 空值兜底为 []
	if (req.Type == "array" || req.Type == "array-json") && (req.Value == "" || req.Value == "null" || req.Value == "undefined") {
		req.Value = "[]"
	}
	// object 空值兜底为 {}
	if req.Type == "object" && (req.Value == "" || req.Value == "null" || req.Value == "undefined") {
		req.Value = "{}"
	}

	// 写入
	_, err = dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
		Group:      req.Group,
		GroupName:  req.GroupName,
		Name:       req.Name,
		Key:        req.Key,
		Value:      req.Value,
		Type:       req.Type,
		Options:    optJSON,
		Rules:      rulesJSON,
		Sort:       req.Sort,
		Remark:     req.Remark,
		CreateTime: utility.NowUnix(),
		UpdateTime: utility.NowUnix(),
	}).Insert()
	if err != nil {
		return nil, err
	}

	// 清理缓存
	// 无缓存模式，创建后无需清理，后续读取直接查库

	res = &api.ConfigCreateRes{}
	return
}

// normalizeValueByType 简单按 type 做格式校验/转换
func normalizeValueByType(t string, v string) (string, error) {
	switch t {
	case "number":
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			return "", err
		}
		return v, nil
	case "switch":
		if v == "true" || v == "1" {
			return "1", nil
		}
		if v == "false" || v == "0" {
			return "0", nil
		}
		return "", gerror.New("开关类型仅支持 0/1/true/false")
	case "json", "object", "array":
		var tmp interface{}
		if err := gjson.DecodeTo(v, &tmp); err != nil {
			return "", err
		}
		// 保持原字符串存储
		return v, nil
	default:
		// text/textarea/select/radio/checkbox/color/upload 等直接存
		return v, nil
	}
}

// getConfigGroup 读取分组配置，带缓存
func getConfigGroup(ctx context.Context, group string) (map[string]string, []entity.SysConfig, error) {
	var items []entity.SysConfig
	if err := dao.SysConfig.Ctx(ctx).
		Where(dao.SysConfig.Columns().Group, group).
		OrderAsc(dao.SysConfig.Columns().Group).OrderAsc("sort").OrderAsc("id").
		Scan(&items); err != nil {
		return nil, nil, err
	}
	if len(items) == 0 {
		return nil, nil, gerror.NewCode(consts.CodeDataNotFound, "分组不存在或无配置项")
	}

	m := make(map[string]string, len(items))
	for _, it := range items {
		m[it.Key] = it.Value
	}

	return m, items, nil
}

// ConfigGroupList 获取配置分组列表
func (c *ControllerV1) ConfigGroupList(ctx context.Context, req *api.ConfigGroupListReq) (res *api.ConfigGroupListRes, err error) {
	// 从 key='config_group' 的配置项中读取
	var configItem entity.SysConfig
	err = dao.SysConfig.Ctx(ctx).Where(dao.SysConfig.Columns().Key, "config_group").Scan(&configItem)
	if err != nil {
		return nil, err
	}

	// 解析 JSON 数组
	var groups []api.ConfigGroupItem
	if configItem.Value != "" {
		if err = gjson.DecodeTo(configItem.Value, &groups); err != nil {
			return nil, gerror.New("解析配置分组数据失败")
		}
	}

	res = &api.ConfigGroupListRes{
		List: groups,
	}
	return
}

// ConfigGroupSave 保存配置分组（添加或编辑）
func (c *ControllerV1) ConfigGroupSave(ctx context.Context, req *api.ConfigGroupSaveReq) (res *api.ConfigGroupSaveRes, err error) {
	// 读取当前配置分组数据
	var configItem entity.SysConfig
	err = dao.SysConfig.Ctx(ctx).Where(dao.SysConfig.Columns().Key, "config_group").Scan(&configItem)
	if err != nil {
		return nil, err
	}

	// 解析现有分组列表
	var groups []api.ConfigGroupItem
	if configItem.Value != "" {
		if err = gjson.DecodeTo(configItem.Value, &groups); err != nil {
			return nil, gerror.New("解析配置分组数据失败")
		}
	}

	newGroup := api.ConfigGroupItem{
		Group:       req.Group,
		GroupName:   req.GroupName,
		Icon:        req.Icon,
		Description: req.Description,
		Sort:        req.Sort,
	}

	if req.IsEdit {
		// 编辑模式：更新已存在的分组
		found := false
		for i, g := range groups {
			if g.Group == req.Group {
				groups[i] = newGroup
				found = true
				break
			}
		}
		if !found {
			return nil, gerror.New("分组不存在")
		}
	} else {
		// 添加模式：检查是否已存在
		for _, g := range groups {
			if g.Group == req.Group {
				return nil, gerror.New("分组标识已存在")
			}
		}
		groups = append(groups, newGroup)
	}

	// 序列化为 JSON
	newValue, err := gjson.Encode(groups)
	if err != nil {
		return nil, gerror.New("序列化配置分组数据失败")
	}

	// 更新数据库
	_, err = dao.SysConfig.Ctx(ctx).
		Data(do.SysConfig{
			Value:      string(newValue),
			UpdateTime: utility.NowUnix(),
		}).
		Where(dao.SysConfig.Columns().Key, "config_group").
		Update()
	if err != nil {
		return nil, err
	}

	res = &api.ConfigGroupSaveRes{}
	return
}

// ConfigGroupDelete 删除配置分组
func (c *ControllerV1) ConfigGroupDelete(ctx context.Context, req *api.ConfigGroupDeleteReq) (res *api.ConfigGroupDeleteRes, err error) {
	// 读取当前配置分组数据
	var configItem entity.SysConfig
	err = dao.SysConfig.Ctx(ctx).Where(dao.SysConfig.Columns().Key, "config_group").Scan(&configItem)
	if err != nil {
		return nil, err
	}

	// 解析现有分组列表
	var groups []api.ConfigGroupItem
	if configItem.Value != "" {
		if err = gjson.DecodeTo(configItem.Value, &groups); err != nil {
			return nil, gerror.New("解析配置分组数据失败")
		}
	}

	// 删除指定分组
	newGroups := make([]api.ConfigGroupItem, 0)
	found := false
	for _, g := range groups {
		if g.Group != req.Group {
			newGroups = append(newGroups, g)
		} else {
			found = true
		}
	}

	if !found {
		return nil, gerror.New("分组不存在")
	}

	// 序列化为 JSON
	newValue, err := gjson.Encode(newGroups)
	if err != nil {
		return nil, gerror.New("序列化配置分组数据失败")
	}

	// 更新数据库
	_, err = dao.SysConfig.Ctx(ctx).
		Data(do.SysConfig{
			Value:      string(newValue),
			UpdateTime: utility.NowUnix(),
		}).
		Where(dao.SysConfig.Columns().Key, "config_group").
		Update()
	if err != nil {
		return nil, err
	}

	// TODO: 可选择是否同时删除该分组下的所有配置项
	// _, err = dao.SysConfig.Ctx(ctx).Where("\"group\" = ?", req.Group).Delete()

	res = &api.ConfigGroupDeleteRes{}
	return
}

// ConfigDelete 删除配置项
func (c *ControllerV1) ConfigDelete(ctx context.Context, req *api.ConfigDeleteReq) (res *api.ConfigDeleteRes, err error) {
	// 查询配置项
	var item entity.SysConfig
	err = dao.SysConfig.Ctx(ctx).Where(dao.SysConfig.Columns().Key, req.Key).Scan(&item)
	if err != nil {
		return nil, err
	}
	if item.Id == 0 {
		return nil, gerror.New("配置项不存在")
	}

	// 检查是否允许删除
	if item.AllowDel == 0 {
		return nil, gerror.New("该配置项不允许删除")
	}

	// 删除配置项
	_, err = dao.SysConfig.Ctx(ctx).Where(dao.SysConfig.Columns().Key, req.Key).Delete()
	if err != nil {
		return nil, err
	}

	res = &api.ConfigDeleteRes{}
	return
}
