-- ============================================================
-- 短信模块菜单 + 按钮权限（PostgreSQL）
-- 挂在「系统管理」目录下，使用 INSERT ... WHERE NOT EXISTS 防重复
-- ============================================================

-- 查找系统管理目录的 ID（name='SystemSetting', type=1）
-- 如果你的项目中系统管理目录 name 不同，请替换

-- 1. 短信模板菜单（type=2）
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 2, '短信模板', 'SmsTemplate', 'sms-template', '/system/sms-template/index', 'sms_template', 'ri:message-2-line', 0, 1, '', '', '["GET /admin/sms/template/list"]', 0, 0, 0, '', '', 0, 0, 80, 1, '短信模板管理', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'System' AND p.type = 1
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu WHERE name = 'SmsTemplate' AND type = 2);

-- 2. 短信变量菜单（type=2）
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 2, '短信变量', 'SmsVariable', 'sms-variable', '/system/sms-variable/index', 'sms_variable', 'ri:braces-line', 0, 1, '', '', '["GET /admin/sms/variable/list"]', 0, 0, 0, '', '', 0, 0, 81, 1, '短信模板变量', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'System' AND p.type = 1
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu WHERE name = 'SmsVariable' AND type = 2);

-- 3. 发送日志菜单（type=2）
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 2, '短信日志', 'SmsLog', 'sms-log', '/system/sms-log/index', 'sms_log', 'ri:file-list-3-line', 0, 1, '', '', '["GET /admin/sms/log/list"]', 0, 0, 0, '', '', 0, 0, 82, 1, '短信发送日志', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'System' AND p.type = 1
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu WHERE name = 'SmsLog' AND type = 2);


-- ========== 按钮权限（type=3） ==========

-- 短信模板 - 新增/编辑
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '新增/编辑', 'edit', '', '', 'sms_template', '', 0, 0, '', '', '["POST /admin/sms/template/save"]', 0, 0, 0, '', '', 0, 0, 1, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'SmsTemplate' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'edit');

-- 短信模板 - 删除
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '删除', 'delete', '', '', 'sms_template', '', 0, 0, '', '', '["POST /admin/sms/template/delete"]', 0, 0, 0, '', '', 0, 0, 2, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'SmsTemplate' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'delete');

-- 短信模板 - 测试发送
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '测试发送', 'test', '', '', 'sms_template', '', 0, 0, '', '', '["POST /admin/sms/template/test"]', 0, 0, 0, '', '', 0, 0, 3, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'SmsTemplate' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'test');

-- 短信变量 - 新增/编辑
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '新增/编辑', 'edit', '', '', 'sms_variable', '', 0, 0, '', '', '["POST /admin/sms/variable/save"]', 0, 0, 0, '', '', 0, 0, 1, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'SmsVariable' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'edit');

-- 短信变量 - 删除
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '删除', 'delete', '', '', 'sms_variable', '', 0, 0, '', '', '["POST /admin/sms/variable/delete"]', 0, 0, 0, '', '', 0, 0, 2, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'SmsVariable' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'delete');
