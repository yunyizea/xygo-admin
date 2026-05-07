-- ============================================================
-- v1.3.4 按钮权限数据升级脚本 —— PostgreSQL 版
-- 通过父菜单 name(路由名) + 按钮 title 定位，不依赖硬编码 ID
-- ============================================================

-- ========== 一、修正 parent_id：会员列表按钮应挂在 MemberList 下 ==========
-- 原始数据中会员按钮的 parent_id 指向 'Member'(type=1 目录)，应改为 'MemberList'(type=2 菜单页)
UPDATE xy_admin_menu SET parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberList' AND type = 2)
WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'Member' AND type = 1)
  AND type = 3;


-- ========== 二、UPDATE 已有 type=3 按钮的 name 为标准 authMark ==========

-- 附件管理 (parent: system/attachment)
UPDATE xy_admin_menu SET name = 'view'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'system/attachment' AND type = 2) AND type = 3 AND title = '查看';
UPDATE xy_admin_menu SET name = 'edit'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'system/attachment' AND type = 2) AND type = 3 AND title = '编辑';
UPDATE xy_admin_menu SET name = 'delete' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'system/attachment' AND type = 2) AND type = 3 AND title = '删除';

-- 会员列表 (parent: MemberList)
UPDATE xy_admin_menu SET name = 'add'      WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberList' AND type = 2) AND type = 3 AND title = '添加会员';
UPDATE xy_admin_menu SET name = 'edit'     WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberList' AND type = 2) AND type = 3 AND title = '编辑会员';
UPDATE xy_admin_menu SET name = 'delete'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberList' AND type = 2) AND type = 3 AND title = '删除会员';
UPDATE xy_admin_menu SET name = 'resetPwd' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberList' AND type = 2) AND type = 3 AND title = '重置密码';

-- 会员分组 (parent: MemberGroup)
UPDATE xy_admin_menu SET name = 'add'    WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberGroup' AND type = 2) AND type = 3 AND title = '新增分组';
UPDATE xy_admin_menu SET name = 'edit'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberGroup' AND type = 2) AND type = 3 AND title = '编辑分组';
UPDATE xy_admin_menu SET name = 'delete' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberGroup' AND type = 2) AND type = 3 AND title = '删除分组';

-- 会员菜单 (parent: MemberMenu)
UPDATE xy_admin_menu SET name = 'add'    WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberMenu' AND type = 2) AND type = 3 AND title = '新增菜单';
UPDATE xy_admin_menu SET name = 'edit'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberMenu' AND type = 2) AND type = 3 AND title = '编辑菜单';
UPDATE xy_admin_menu SET name = 'delete' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberMenu' AND type = 2) AND type = 3 AND title = '删除菜单';

-- 登录日志-安全监控 (parent: LoginLog)
UPDATE xy_admin_menu SET name = 'batchDel' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'LoginLog' AND type = 2) AND type = 3 AND title = '删除日志';
UPDATE xy_admin_menu SET name = 'clear'    WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'LoginLog' AND type = 2) AND type = 3 AND title = '清空日志';

-- 操作日志 (parent: OperationLog)
UPDATE xy_admin_menu SET name = 'detail' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'OperationLog' AND type = 2) AND type = 3 AND title = '查看详情';
UPDATE xy_admin_menu SET name = 'delete' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'OperationLog' AND type = 2) AND type = 3 AND title = '删除日志';
UPDATE xy_admin_menu SET name = 'clear'  WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'OperationLog' AND type = 2) AND type = 3 AND title = '清空日志';

-- 通知管理 (parent: Notice)
UPDATE xy_admin_menu SET name = 'view'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'Notice' AND type = 2) AND type = 3 AND title = '查看';
UPDATE xy_admin_menu SET name = 'edit'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'Notice' AND type = 2) AND type = 3 AND title = '发布/编辑';
UPDATE xy_admin_menu SET name = 'delete' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'Notice' AND type = 2) AND type = 3 AND title = '删除';

-- 定时任务 (parent: CronManage)
UPDATE xy_admin_menu SET name = 'view'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'CronManage' AND type = 2) AND type = 3 AND title = '查看';
UPDATE xy_admin_menu SET name = 'edit'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'CronManage' AND type = 2) AND type = 3 AND title = '新增/编辑';
UPDATE xy_admin_menu SET name = 'delete' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'CronManage' AND type = 2) AND type = 3 AND title = '删除';
UPDATE xy_admin_menu SET name = 'exec'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'CronManage' AND type = 2) AND type = 3 AND title = '在线执行';

-- 会员登录日志 (parent: MemberLoginLog)
UPDATE xy_admin_menu SET name = 'view'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberLoginLog' AND type = 2) AND type = 3 AND title = '查看登录日志';
UPDATE xy_admin_menu SET name = 'delete' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberLoginLog' AND type = 2) AND type = 3 AND title = '删除登录日志';
UPDATE xy_admin_menu SET name = 'export' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberLoginLog' AND type = 2) AND type = 3 AND title = '导出登录日志';

-- CMS 文档 (parent: CmsDoc)
UPDATE xy_admin_menu SET name = 'edit'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'CmsDoc' AND type = 2) AND type = 3 AND title = '新增/编辑';
UPDATE xy_admin_menu SET name = 'delete' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'CmsDoc' AND type = 2) AND type = 3 AND title = '删除';

-- CMS 文档分类 (parent: CmsDocCategory)
UPDATE xy_admin_menu SET name = 'edit'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'CmsDocCategory' AND type = 2) AND type = 3 AND title = '新增/编辑';
UPDATE xy_admin_menu SET name = 'delete' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'CmsDocCategory' AND type = 2) AND type = 3 AND title = '删除';

-- 余额变动日志 (parent: MemberMoneyLog)
UPDATE xy_admin_menu SET name = 'view'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberMoneyLog' AND type = 2) AND type = 3 AND title LIKE '%查看%';
UPDATE xy_admin_menu SET name = 'add'    WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberMoneyLog' AND type = 2) AND type = 3 AND title LIKE '%新增%';
UPDATE xy_admin_menu SET name = 'edit'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberMoneyLog' AND type = 2) AND type = 3 AND title LIKE '%编辑%';
UPDATE xy_admin_menu SET name = 'delete' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberMoneyLog' AND type = 2) AND type = 3 AND title LIKE '%删除%';
UPDATE xy_admin_menu SET name = 'export' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberMoneyLog' AND type = 2) AND type = 3 AND title LIKE '%导出%';

-- 会员通知 (parent: MemberNotice)
UPDATE xy_admin_menu SET name = 'view'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberNotice' AND type = 2) AND type = 3 AND title LIKE '%查看%';
UPDATE xy_admin_menu SET name = 'add'    WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberNotice' AND type = 2) AND type = 3 AND title LIKE '%新增%';
UPDATE xy_admin_menu SET name = 'edit'   WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberNotice' AND type = 2) AND type = 3 AND title LIKE '%编辑%';
UPDATE xy_admin_menu SET name = 'delete' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberNotice' AND type = 2) AND type = 3 AND title LIKE '%删除%';
UPDATE xy_admin_menu SET name = 'export' WHERE parent_id = (SELECT id FROM xy_admin_menu WHERE name = 'MemberNotice' AND type = 2) AND type = 3 AND title LIKE '%导出%';


-- ========== 三、INSERT 缺失的 type=3 按钮（不指定 id，自增生成） ==========
-- 使用 INSERT ... SELECT ... WHERE NOT EXISTS 防止重复执行

-- 后台用户 (parent: User)
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '新增用户', 'add', '', '', 'admin_user', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 1, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'User' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'add');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '编辑用户', 'edit', '', '', 'admin_user', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 2, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'User' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'edit');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '删除用户', 'delete', '', '', 'admin_user', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 3, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'User' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'delete');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '强制下线', 'kick', '', '', 'admin_user', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 4, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'User' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'kick');

-- 角色管理 (parent: Role)
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '新增角色', 'add', '', '', 'admin_role', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 1, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'Role' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'add');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '编辑角色', 'edit', '', '', 'admin_role', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 2, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'Role' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'edit');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '删除角色', 'delete', '', '', 'admin_role', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 3, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'Role' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'delete');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '菜单权限', 'permission', '', '', 'admin_role', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 4, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'Role' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'permission');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '数据权限', 'dataScope', '', '', 'admin_role', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 5, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'Role' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'dataScope');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '字段权限', 'fieldPerm', '', '', 'admin_role', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 6, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'Role' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'fieldPerm');

-- 部门管理 (parent: Dept)
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '添加部门', 'add', '', '', 'admin_dept', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 1, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'Dept' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'add');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '编辑部门', 'edit', '', '', 'admin_dept', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 2, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'Dept' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'edit');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '删除部门', 'delete', '', '', 'admin_dept', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 3, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'Dept' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'delete');

-- 岗位管理 (parent: Post)
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '新增岗位', 'add', '', '', 'admin_post', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 1, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'Post' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'add');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '编辑岗位', 'edit', '', '', 'admin_post', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 2, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'Post' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'edit');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '删除岗位', 'delete', '', '', 'admin_post', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 3, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'Post' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'delete');

-- 附件管理 - 补 add, batchDel (parent: system/attachment)
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '上传文件', 'add', '', '', 'sys_attachment', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 0, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'system/attachment' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'add');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '批量删除', 'batchDel', '', '', 'sys_attachment', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 4, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'system/attachment' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'batchDel');

-- 通知管理 - 补 add (parent: Notice)
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '发布通知', 'add', '', '', 'admin_notice', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 0, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'Notice' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'add');

-- 定时任务 - 补 add (parent: CronManage)
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '新增任务', 'add', '', '', 'sys_cron', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 0, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'CronManage' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'add');

-- 会员列表 - 补 batchDel (parent: MemberList)
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '批量删除', 'batchDel', '', '', '', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 5, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'MemberList' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'batchDel');

-- 操作日志 - 补 batchDel (parent: OperationLog)
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '批量删除', 'batchDel', '', '', 'admin_operation_log', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 0, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'OperationLog' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'batchDel');

-- 会员登录日志 - 补 batchDel (parent: MemberLoginLog)
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '批量删除', 'batchDel', '', '', '', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 2, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'MemberLoginLog' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'batchDel');

-- 积分变动日志 - 全部缺失 (parent: MemberScoreLog)
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '新增', 'add', '', '', '', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 1, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'MemberScoreLog' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'add');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '查看', 'view', '', '', '', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 2, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'MemberScoreLog' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'view');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '编辑', 'edit', '', '', '', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 3, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'MemberScoreLog' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'edit');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '删除', 'delete', '', '', '', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 4, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'MemberScoreLog' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'delete');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '批量删除', 'batchDel', '', '', '', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 5, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'MemberScoreLog' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'batchDel');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '导出', 'export', '', '', '', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 6, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'MemberScoreLog' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'export');

-- 余额变动日志 - 补 batchDel (parent: MemberMoneyLog)
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '批量删除', 'batchDel', '', '', '', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 6, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'MemberMoneyLog' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'batchDel');

-- 会员通知 - 补 batchDel (parent: MemberNotice)
INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '批量删除', 'batchDel', '', '', '', '', 0, 0, '', '', '', 0, 0, 0, '', '', 0, 0, 6, 1, '', 0, 0, EXTRACT(EPOCH FROM NOW())::bigint, EXTRACT(EPOCH FROM NOW())::bigint
FROM xy_admin_menu p WHERE p.name = 'MemberNotice' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'batchDel');
