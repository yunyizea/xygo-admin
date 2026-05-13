-- Migration: 1.3.8
-- Description: 新增短信模块表、配置、菜单和按钮权限

CREATE TABLE IF NOT EXISTS `xy_sms_template` (
    `id`                   bigint unsigned NOT NULL AUTO_INCREMENT,
    `title`                varchar(128) NOT NULL DEFAULT '' COMMENT '模板标题',
    `code`                 varchar(64)  NOT NULL DEFAULT '' COMMENT '模板唯一标识（如 user_register）',
    `content`              text          NOT NULL COMMENT '短信文案（含变量占位 ${var}）',
    `provider_template_id` varchar(64)  NOT NULL DEFAULT '' COMMENT '服务商模板ID',
    `variables`            json          DEFAULT NULL COMMENT '模板变量列表 JSON',
    `related_variable_id`  bigint unsigned NOT NULL DEFAULT 0 COMMENT '关联文案变量ID',
    `status`               tinyint       NOT NULL DEFAULT 1 COMMENT '状态：1=启用 0=禁用',
    `sort`                 int           NOT NULL DEFAULT 0 COMMENT '排序',
    `remark`               varchar(255)  NOT NULL DEFAULT '' COMMENT '备注',
    `created_by`           bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建人ID',
    `updated_by`           bigint unsigned NOT NULL DEFAULT 0 COMMENT '更新人ID',
    `create_time`          bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间（Unix秒）',
    `update_time`          bigint unsigned NOT NULL DEFAULT 0 COMMENT '更新时间（Unix秒）',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_sms_template_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='短信模板';

CREATE TABLE IF NOT EXISTS `xy_sms_variable` (
    `id`            bigint unsigned NOT NULL AUTO_INCREMENT,
    `title`         varchar(128) NOT NULL DEFAULT '' COMMENT '变量标题',
    `name`          varchar(64)  NOT NULL DEFAULT '' COMMENT '变量名（如 usermobile）',
    `source_type`   tinyint      NOT NULL DEFAULT 1 COMMENT '来源类型：1=字段提取 2=SQL查询 3=内置Helper',
    `sql_query`     text         NOT NULL COMMENT 'SQL查询语句（source_type=2 时）',
    `method_name`   varchar(128) NOT NULL DEFAULT '' COMMENT 'Helper方法路径（source_type=3 时）',
    `shared_count`  int          NOT NULL DEFAULT 0 COMMENT '共通数据数',
    `status`        tinyint      NOT NULL DEFAULT 1 COMMENT '状态：1=启用 0=禁用',
    `created_by`    bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建人ID',
    `updated_by`    bigint unsigned NOT NULL DEFAULT 0 COMMENT '更新人ID',
    `create_time`   bigint unsigned NOT NULL DEFAULT 0 COMMENT '创建时间（Unix秒）',
    `update_time`   bigint unsigned NOT NULL DEFAULT 0 COMMENT '更新时间（Unix秒）',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_sms_variable_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='短信模板变量';

CREATE TABLE IF NOT EXISTS `xy_sms_log` (
    `id`             bigint unsigned NOT NULL AUTO_INCREMENT,
    `phone`          varchar(20)  NOT NULL DEFAULT '' COMMENT '手机号',
    `template_code`  varchar(64)  NOT NULL DEFAULT '' COMMENT '使用的模板标识',
    `driver`         varchar(32)  NOT NULL DEFAULT '' COMMENT '驱动名（aliyun/tencent）',
    `content`        text         NOT NULL COMMENT '实际发送内容',
    `params`         json         DEFAULT NULL COMMENT '发送参数 JSON',
    `status`         tinyint      NOT NULL DEFAULT 0 COMMENT '状态：1=成功 0=失败',
    `request_id`     varchar(128) NOT NULL DEFAULT '' COMMENT '服务商返回请求ID',
    `error_msg`      text         NOT NULL COMMENT '错误信息',
    `create_time`    bigint unsigned NOT NULL DEFAULT 0 COMMENT '发送时间（Unix秒）',
    PRIMARY KEY (`id`),
    KEY `idx_sms_log_phone` (`phone`),
    KEY `idx_sms_log_template_code` (`template_code`),
    KEY `idx_sms_log_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='短信发送日志';

UPDATE xy_sys_config
SET value = JSON_ARRAY_APPEND(value, '$', JSON_OBJECT('group','sms','groupName','短信配置','icon','ri:smartphone-line','description','配置短信接口','sort',40)),
    update_time = UNIX_TIMESTAMP()
WHERE `key` = 'config_group'
  AND JSON_VALID(value)
  AND NOT JSON_CONTAINS(value, '{"group":"sms"}');

INSERT IGNORE INTO `xy_sys_config` (`group`, `group_name`, `name`, `key`, `value`, `type`, `options`, `rules`, `sort`, `remark`, `allow_del`, `created_by`, `updated_by`, `create_time`, `update_time`)
VALUES
('sms', '短信配置', '发送超时（秒）', 'sms_timeout', '5', 'number', NULL, '{"required": true}', 10, '短信发送超时时间', 0, 0, 0, 0, 0),
('sms', '短信配置', '发送策略', 'sms_strategy', 'weight', 'select', '{"options": [{"label": "按权重", "value": "weight"}, {"label": "随机", "value": "random"}]}', '{"required": true}', 20, '多驱动时的选择策略', 0, 0, 0, 0, 0),
('sms', '短信配置', '启用的服务商', 'sms_enabled_drivers', '', 'selects', '{"options": [{"label": "阿里云", "value": "aliyun"}, {"label": "腾讯云", "value": "tencent"}]}', NULL, 30, '可多选，逗号分隔', 0, 0, 0, 0, 0),
('sms', '短信配置', '阿里云 AccessKey ID', 'sms_aliyun_access_key_id', '', 'text', NULL, NULL, 100, '', 0, 0, 0, 0, 0),
('sms', '短信配置', '阿里云 AccessKey Secret', 'sms_aliyun_access_key_secret', '', 'password', NULL, NULL, 110, '', 0, 0, 0, 0, 0),
('sms', '短信配置', '阿里云短信签名', 'sms_aliyun_sign_name', '', 'text', NULL, NULL, 120, '在阿里云控制台申请的签名', 0, 0, 0, 0, 0),
('sms', '短信配置', '腾讯云 SecretId', 'sms_tencent_secret_id', '', 'text', NULL, NULL, 200, '', 0, 0, 0, 0, 0),
('sms', '短信配置', '腾讯云 SecretKey', 'sms_tencent_secret_key', '', 'password', NULL, NULL, 210, '', 0, 0, 0, 0, 0),
('sms', '短信配置', '腾讯云 AppId', 'sms_tencent_app_id', '', 'text', NULL, NULL, 220, '腾讯云短信应用 SDK AppID', 0, 0, 0, 0, 0),
('sms', '短信配置', '腾讯云短信签名', 'sms_tencent_sign_name', '', 'text', NULL, NULL, 230, '在腾讯云控制台申请的签名', 0, 0, 0, 0, 0);

INSERT IGNORE INTO `xy_sms_template` (`title`, `code`, `content`, `provider_template_id`, `variables`, `status`, `sort`, `remark`, `created_by`, `updated_by`, `create_time`, `update_time`)
VALUES
('用户注册', 'user_register', '{1}为您的登录验证码，请于{2}分钟内填写，如非本人操作，请忽略本短信。', '2627353', '["code", "expire_minutes"]', 1, 10, '', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('登录验证', 'user_login', '{1}为您的登录验证码，请于{2}分钟内填写，如非本人操作，请忽略本短信。', '2627353', '["code", "expire_minutes"]', 1, 20, '', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('修改密码', 'reset_password', '【${site_name}】您正在修改密码，验证码为：${code}，有效期${expire_minutes}分钟。如非本人操作，请立即修改密码。', '', '["code", "expire_minutes", "site_name"]', 1, 30, '', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('订单通知', 'order_notify', '【${site_name}】您的订单已提交成功，请保持手机畅通。', '', '["site_name"]', 1, 40, '', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT IGNORE INTO `xy_sms_variable` (`title`, `name`, `source_type`, `sql_query`, `method_name`, `shared_count`, `status`, `created_by`, `updated_by`, `create_time`, `update_time`)
VALUES
('验证码', 'code', 1, '', '', 0, 1, 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('手机号', 'mobile', 1, '', '', 0, 1, 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('平台名称', 'site_name', 1, '', '', 0, 1, 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('过期时间(分钟)', 'expire_minutes', 1, '', '', 0, 1, 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 2, '短信模板', 'SmsTemplate', 'sms-template', '/system/sms-template/index', 'sms_template', 'ri:message-2-line', 0, 1, '', '', '["GET /admin/sms/template/list"]', 0, 0, 0, '', '', 0, 0, 80, 1, '短信模板管理', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
FROM xy_admin_menu p WHERE p.name = 'System' AND p.type = 1
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu t WHERE t.name = 'SmsTemplate' AND t.type = 2);

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 2, '短信变量', 'SmsVariable', 'sms-variable', '/system/sms-variable/index', 'sms_variable', 'ri:braces-line', 0, 1, '', '', '["GET /admin/sms/variable/list"]', 0, 0, 0, '', '', 0, 0, 81, 1, '短信模板变量', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
FROM xy_admin_menu p WHERE p.name = 'System' AND p.type = 1
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu t WHERE t.name = 'SmsVariable' AND t.type = 2);

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 2, '短信日志', 'SmsLog', 'sms-log', '/system/sms-log/index', 'sms_log', 'ri:file-list-3-line', 0, 1, '', '', '["GET /admin/sms/log/list"]', 0, 0, 0, '', '', 0, 0, 82, 1, '短信发送日志', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
FROM xy_admin_menu p WHERE p.name = 'System' AND p.type = 1
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu t WHERE t.name = 'SmsLog' AND t.type = 2);

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '新增/编辑', 'edit', '', '', 'sms_template', '', 0, 0, '', '', '["POST /admin/sms/template/save"]', 0, 0, 0, '', '', 0, 0, 1, 1, '', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
FROM xy_admin_menu p WHERE p.name = 'SmsTemplate' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'edit');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '删除', 'delete', '', '', 'sms_template', '', 0, 0, '', '', '["POST /admin/sms/template/delete"]', 0, 0, 0, '', '', 0, 0, 2, 1, '', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
FROM xy_admin_menu p WHERE p.name = 'SmsTemplate' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'delete');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '测试发送', 'test', '', '', 'sms_template', '', 0, 0, '', '', '["POST /admin/sms/template/test"]', 0, 0, 0, '', '', 0, 0, 3, 1, '', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
FROM xy_admin_menu p WHERE p.name = 'SmsTemplate' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'test');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '新增/编辑', 'edit', '', '', 'sms_variable', '', 0, 0, '', '', '["POST /admin/sms/variable/save"]', 0, 0, 0, '', '', 0, 0, 1, 1, '', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
FROM xy_admin_menu p WHERE p.name = 'SmsVariable' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'edit');

INSERT INTO xy_admin_menu (parent_id, type, title, name, path, component, resource, icon, hidden, keep_alive, redirect, frame_src, perms, is_frame, affix, show_badge, badge_text, active_path, hide_tab, is_full_page, sort, status, remark, created_by, updated_by, create_time, update_time)
SELECT p.id, 3, '删除', 'delete', '', '', 'sms_variable', '', 0, 0, '', '', '["POST /admin/sms/variable/delete"]', 0, 0, 0, '', '', 0, 0, 2, 1, '', 0, 0, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()
FROM xy_admin_menu p WHERE p.name = 'SmsVariable' AND p.type = 2
AND NOT EXISTS (SELECT 1 FROM xy_admin_menu sub WHERE sub.parent_id = p.id AND sub.type = 3 AND sub.name = 'delete');
