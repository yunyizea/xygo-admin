-- {{.TableComment}} 菜单 SQL
-- 上级菜单ID: {{.MenuPid}}

{{- if eq .MenuPid 0}}
-- ======= 顶级模式：创建目录(type=1) + 页面(type=2) + 按钮(type=3) =======

-- 1. 创建目录
INSERT INTO `xy_admin_menu` (`parent_id`, `type`, `title`, `name`, `path`, `component`, `resource`, `icon`, `hidden`, `keep_alive`, `redirect`, `frame_src`, `perms`, `is_frame`, `affix`, `show_badge`, `badge_text`, `active_path`, `hide_tab`, `is_full_page`, `sort`, `status`, `remark`, `created_by`, `updated_by`, `create_time`, `update_time`)
VALUES (0, 1, '{{.TableComment}}', '{{.VarName}}Dir', '/{{.RouteName}}', '', '', '{{.MenuIcon}}', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, {{.MenuSort}}, 1, '', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

SET @parentId = LAST_INSERT_ID();

-- 2. 创建页面菜单
INSERT INTO `xy_admin_menu` (`parent_id`, `type`, `title`, `name`, `path`, `component`, `resource`, `icon`, `hidden`, `keep_alive`, `redirect`, `frame_src`, `perms`, `is_frame`, `affix`, `show_badge`, `badge_text`, `active_path`, `hide_tab`, `is_full_page`, `sort`, `status`, `remark`, `created_by`, `updated_by`, `create_time`, `update_time`)
VALUES (@parentId, 2, '{{.TableComment}}列表', '{{.VarName}}', '{{.RouteName}}', '{{.MenuComponentPath}}/index', '{{.ResourceName}}', '', 0, 1, '', '', '["GET {{.ApiPrefix}}/list"]', 0, 0, 0, '', '', 0, 0, 1, 1, '{{.MenuRemark}}', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

SET @pageId = LAST_INSERT_ID();

{{- else}}
-- ======= 挂载模式：在已有目录下创建页面(type=2) + 按钮(type=3) =======

-- 1. 创建页面菜单（挂载到上级 #{{.MenuPid}}）
INSERT INTO `xy_admin_menu` (`parent_id`, `type`, `title`, `name`, `path`, `component`, `resource`, `icon`, `hidden`, `keep_alive`, `redirect`, `frame_src`, `perms`, `is_frame`, `affix`, `show_badge`, `badge_text`, `active_path`, `hide_tab`, `is_full_page`, `sort`, `status`, `remark`, `created_by`, `updated_by`, `create_time`, `update_time`)
VALUES ({{.MenuPid}}, 2, '{{.TableComment}}', '{{.VarName}}', '{{.RouteName}}', '{{.MenuComponentPath}}/index', '{{.ResourceName}}', '{{.MenuIcon}}', 0, 1, '', '', '["GET {{.ApiPrefix}}/list"]', 0, 0, 0, '', '', 0, 0, {{.MenuSort}}, 1, '{{.MenuRemark}}', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

SET @pageId = LAST_INSERT_ID();

{{- end}}

-- 3. 创建按钮权限（根据选项按需生成）
{{- $btnIdx := 1}}
{{- if .HasView}}
INSERT INTO `xy_admin_menu` (`parent_id`, `type`, `title`, `name`, `path`, `component`, `resource`, `icon`, `hidden`, `keep_alive`, `redirect`, `frame_src`, `perms`, `is_frame`, `affix`, `show_badge`, `badge_text`, `active_path`, `hide_tab`, `is_full_page`, `sort`, `status`, `remark`, `created_by`, `updated_by`, `create_time`, `update_time`)
VALUES (@pageId, 3, '查看{{.TableComment}}', '{{.VarName}}View', '', '', '', '', 0, 0, '', '', '["GET {{.ApiPrefix}}/view"]', 0, 0, 0, '', '', 0, 0, 1, 1, '{{.MenuRemark}}', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());
{{- if eq .ViewMode "page"}}
-- 详情页路由（隐藏页面，与列表页同级，active_path 高亮列表页）
{{- if eq .MenuPid 0}}
INSERT INTO `xy_admin_menu` (`parent_id`, `type`, `title`, `name`, `path`, `component`, `resource`, `icon`, `hidden`, `keep_alive`, `redirect`, `frame_src`, `perms`, `is_frame`, `affix`, `show_badge`, `badge_text`, `active_path`, `hide_tab`, `is_full_page`, `sort`, `status`, `remark`, `created_by`, `updated_by`, `create_time`, `update_time`)
VALUES (@parentId, 2, '{{.TableComment}}详情', '{{.VarName}}Detail', '{{.RouteName}}/detail', '{{.MenuComponentPath}}/detail/index', '', '', 1, 0, '', '', '["GET {{.ApiPrefix}}/view"]', 0, 0, 0, '', '/{{.RouteName}}', 0, 0, 0, 1, '{{.MenuRemark}}', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());
{{- else}}
INSERT INTO `xy_admin_menu` (`parent_id`, `type`, `title`, `name`, `path`, `component`, `resource`, `icon`, `hidden`, `keep_alive`, `redirect`, `frame_src`, `perms`, `is_frame`, `affix`, `show_badge`, `badge_text`, `active_path`, `hide_tab`, `is_full_page`, `sort`, `status`, `remark`, `created_by`, `updated_by`, `create_time`, `update_time`)
VALUES ({{.MenuPid}}, 2, '{{.TableComment}}详情', '{{.VarName}}Detail', '{{.RouteName}}/detail', '{{.MenuComponentPath}}/detail/index', '', '', 1, 0, '', '', '["GET {{.ApiPrefix}}/view"]', 0, 0, 0, '', '/{{.RouteName}}', 0, 0, 0, 1, '{{.MenuRemark}}', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());
{{- end}}
{{- end}}
{{- end}}
{{- if .HasAdd}}
INSERT INTO `xy_admin_menu` (`parent_id`, `type`, `title`, `name`, `path`, `component`, `resource`, `icon`, `hidden`, `keep_alive`, `redirect`, `frame_src`, `perms`, `is_frame`, `affix`, `show_badge`, `badge_text`, `active_path`, `hide_tab`, `is_full_page`, `sort`, `status`, `remark`, `created_by`, `updated_by`, `create_time`, `update_time`)
VALUES (@pageId, 3, '新增{{.TableComment}}', '{{.VarName}}Add', '', '', '', '', 0, 0, '', '', '["POST {{.ApiPrefix}}/edit"]', 0, 0, 0, '', '', 0, 0, 2, 1, '{{.MenuRemark}}', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());
{{- end}}
{{- if .HasEdit}}
INSERT INTO `xy_admin_menu` (`parent_id`, `type`, `title`, `name`, `path`, `component`, `resource`, `icon`, `hidden`, `keep_alive`, `redirect`, `frame_src`, `perms`, `is_frame`, `affix`, `show_badge`, `badge_text`, `active_path`, `hide_tab`, `is_full_page`, `sort`, `status`, `remark`, `created_by`, `updated_by`, `create_time`, `update_time`)
VALUES (@pageId, 3, '编辑{{.TableComment}}', '{{.VarName}}Edit', '', '', '', '', 0, 0, '', '', '["POST {{.ApiPrefix}}/edit","GET {{.ApiPrefix}}/view"]', 0, 0, 0, '', '', 0, 0, 3, 1, '{{.MenuRemark}}', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());
{{- end}}
{{- if or .HasDel .HasBatchDel}}
INSERT INTO `xy_admin_menu` (`parent_id`, `type`, `title`, `name`, `path`, `component`, `resource`, `icon`, `hidden`, `keep_alive`, `redirect`, `frame_src`, `perms`, `is_frame`, `affix`, `show_badge`, `badge_text`, `active_path`, `hide_tab`, `is_full_page`, `sort`, `status`, `remark`, `created_by`, `updated_by`, `create_time`, `update_time`)
VALUES (@pageId, 3, '删除{{.TableComment}}', '{{.VarName}}Delete', '', '', '', '', 0, 0, '', '', '["POST {{.ApiPrefix}}/delete"]', 0, 0, 0, '', '', 0, 0, 4, 1, '{{.MenuRemark}}', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());
{{- end}}
{{- if .HasExport}}
INSERT INTO `xy_admin_menu` (`parent_id`, `type`, `title`, `name`, `path`, `component`, `resource`, `icon`, `hidden`, `keep_alive`, `redirect`, `frame_src`, `perms`, `is_frame`, `affix`, `show_badge`, `badge_text`, `active_path`, `hide_tab`, `is_full_page`, `sort`, `status`, `remark`, `created_by`, `updated_by`, `create_time`, `update_time`)
VALUES (@pageId, 3, '导出{{.TableComment}}', '{{.VarName}}Export', '', '', '', '', 0, 0, '', '', '["GET {{.ApiPrefix}}/export"]', 0, 0, 0, '', '', 0, 0, 5, 1, '{{.MenuRemark}}', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());
{{- end}}
