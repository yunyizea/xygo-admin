-- ============================================================
-- 短信模块建表脚本（PostgreSQL）
-- 表前缀：xy_sms_
-- ============================================================

-- 短信模板
CREATE TABLE IF NOT EXISTS xy_sms_template (
    id                   bigserial PRIMARY KEY,
    title                character varying(128) DEFAULT '' NOT NULL,
    code                 character varying(64) DEFAULT '' NOT NULL,
    content              text DEFAULT '' NOT NULL,
    provider_template_id character varying(64) DEFAULT '' NOT NULL,
    variables            jsonb DEFAULT '[]'::jsonb,
    related_variable_id  bigint DEFAULT 0 NOT NULL,
    status               smallint DEFAULT 1 NOT NULL,
    sort                 integer DEFAULT 0 NOT NULL,
    remark               character varying(255) DEFAULT '' NOT NULL,
    created_by           bigint DEFAULT 0 NOT NULL,
    updated_by           bigint DEFAULT 0 NOT NULL,
    create_time          bigint DEFAULT 0 NOT NULL,
    update_time          bigint DEFAULT 0 NOT NULL
);

COMMENT ON TABLE  xy_sms_template IS '短信模板';
COMMENT ON COLUMN xy_sms_template.id IS '主键';
COMMENT ON COLUMN xy_sms_template.title IS '模板标题';
COMMENT ON COLUMN xy_sms_template.code IS '模板唯一标识（如 user_register）';
COMMENT ON COLUMN xy_sms_template.content IS '短信文案（含变量占位 ${var}）';
COMMENT ON COLUMN xy_sms_template.provider_template_id IS '服务商模板ID';
COMMENT ON COLUMN xy_sms_template.variables IS '模板变量列表 JSON';
COMMENT ON COLUMN xy_sms_template.related_variable_id IS '关联文案变量ID';
COMMENT ON COLUMN xy_sms_template.status IS '状态：1=启用 0=禁用';
COMMENT ON COLUMN xy_sms_template.sort IS '排序';
COMMENT ON COLUMN xy_sms_template.remark IS '备注';
COMMENT ON COLUMN xy_sms_template.created_by IS '创建人ID';
COMMENT ON COLUMN xy_sms_template.updated_by IS '更新人ID';
COMMENT ON COLUMN xy_sms_template.create_time IS '创建时间（Unix秒）';
COMMENT ON COLUMN xy_sms_template.update_time IS '更新时间（Unix秒）';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sms_template_code ON xy_sms_template (code);


-- 短信变量
CREATE TABLE IF NOT EXISTS xy_sms_variable (
    id            bigserial PRIMARY KEY,
    title         character varying(128) DEFAULT '' NOT NULL,
    name          character varying(64) DEFAULT '' NOT NULL,
    source_type   smallint DEFAULT 1 NOT NULL,
    sql_query     text DEFAULT '' NOT NULL,
    method_name   character varying(128) DEFAULT '' NOT NULL,
    shared_count  integer DEFAULT 0 NOT NULL,
    status        smallint DEFAULT 1 NOT NULL,
    created_by    bigint DEFAULT 0 NOT NULL,
    updated_by    bigint DEFAULT 0 NOT NULL,
    create_time   bigint DEFAULT 0 NOT NULL,
    update_time   bigint DEFAULT 0 NOT NULL
);

COMMENT ON TABLE  xy_sms_variable IS '短信模板变量';
COMMENT ON COLUMN xy_sms_variable.id IS '主键';
COMMENT ON COLUMN xy_sms_variable.title IS '变量标题';
COMMENT ON COLUMN xy_sms_variable.name IS '变量名（如 usermobile）';
COMMENT ON COLUMN xy_sms_variable.source_type IS '来源类型：1=字段提取 2=SQL查询 3=内置Helper';
COMMENT ON COLUMN xy_sms_variable.sql_query IS 'SQL查询语句（source_type=2 时）';
COMMENT ON COLUMN xy_sms_variable.method_name IS 'Helper方法路径（source_type=3 时）';
COMMENT ON COLUMN xy_sms_variable.shared_count IS '共通数据数';
COMMENT ON COLUMN xy_sms_variable.status IS '状态：1=启用 0=禁用';
COMMENT ON COLUMN xy_sms_variable.created_by IS '创建人ID';
COMMENT ON COLUMN xy_sms_variable.updated_by IS '更新人ID';
COMMENT ON COLUMN xy_sms_variable.create_time IS '创建时间（Unix秒）';
COMMENT ON COLUMN xy_sms_variable.update_time IS '更新时间（Unix秒）';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sms_variable_name ON xy_sms_variable (name);


-- 短信发送日志
CREATE TABLE IF NOT EXISTS xy_sms_log (
    id             bigserial PRIMARY KEY,
    phone          character varying(20) DEFAULT '' NOT NULL,
    template_code  character varying(64) DEFAULT '' NOT NULL,
    driver         character varying(32) DEFAULT '' NOT NULL,
    content        text DEFAULT '' NOT NULL,
    params         jsonb DEFAULT '{}'::jsonb,
    status         smallint DEFAULT 0 NOT NULL,
    request_id     character varying(128) DEFAULT '' NOT NULL,
    error_msg      text DEFAULT '' NOT NULL,
    create_time    bigint DEFAULT 0 NOT NULL
);

COMMENT ON TABLE  xy_sms_log IS '短信发送日志';
COMMENT ON COLUMN xy_sms_log.id IS '主键';
COMMENT ON COLUMN xy_sms_log.phone IS '手机号';
COMMENT ON COLUMN xy_sms_log.template_code IS '使用的模板标识';
COMMENT ON COLUMN xy_sms_log.driver IS '驱动名（aliyun/tencent）';
COMMENT ON COLUMN xy_sms_log.content IS '实际发送内容';
COMMENT ON COLUMN xy_sms_log.params IS '发送参数 JSON';
COMMENT ON COLUMN xy_sms_log.status IS '状态：1=成功 0=失败';
COMMENT ON COLUMN xy_sms_log.request_id IS '服务商返回请求ID';
COMMENT ON COLUMN xy_sms_log.error_msg IS '错误信息';
COMMENT ON COLUMN xy_sms_log.create_time IS '发送时间（Unix秒）';

CREATE INDEX IF NOT EXISTS idx_sms_log_phone ON xy_sms_log (phone);
CREATE INDEX IF NOT EXISTS idx_sms_log_template_code ON xy_sms_log (template_code);
CREATE INDEX IF NOT EXISTS idx_sms_log_create_time ON xy_sms_log (create_time);


-- ============================================================
-- sys_config 配置数据（短信配置组）
-- ============================================================
INSERT INTO xy_sys_config ("group", group_name, name, key, value, type, options, rules, sort, remark, allow_del, created_by, updated_by, create_time, update_time)
VALUES
('sms', '短信配置', '发送超时（秒）', 'sms_timeout', '5', 'number', NULL, '{"required": true}', 10, '短信发送超时时间', 0, 0, 0, 0, 0),
('sms', '短信配置', '发送策略', 'sms_strategy', 'weight', 'select', '{"options": [{"label": "按权重", "value": "weight"}, {"label": "随机", "value": "random"}]}', '{"required": true}', 20, '多驱动时的选择策略', 0, 0, 0, 0, 0),
('sms', '短信配置', '启用的服务商', 'sms_enabled_drivers', '', 'selects', '{"options": [{"label": "阿里云", "value": "aliyun"}, {"label": "腾讯云", "value": "tencent"}]}', NULL, 30, '可多选，逗号分隔', 0, 0, 0, 0, 0),
('sms', '短信配置', '阿里云 AccessKey ID', 'sms_aliyun_access_key_id', '', 'text', NULL, NULL, 100, '', 0, 0, 0, 0, 0),
('sms', '短信配置', '阿里云 AccessKey Secret', 'sms_aliyun_access_key_secret', '', 'text', NULL, NULL, 110, '', 0, 0, 0, 0, 0),
('sms', '短信配置', '阿里云短信签名', 'sms_aliyun_sign_name', '', 'text', NULL, NULL, 120, '在阿里云控制台申请的签名', 0, 0, 0, 0, 0),
('sms', '短信配置', '腾讯云 SecretId', 'sms_tencent_secret_id', '', 'text', NULL, NULL, 200, '', 0, 0, 0, 0, 0),
('sms', '短信配置', '腾讯云 SecretKey', 'sms_tencent_secret_key', '', 'text', NULL, NULL, 210, '', 0, 0, 0, 0, 0),
('sms', '短信配置', '腾讯云 AppId', 'sms_tencent_app_id', '', 'text', NULL, NULL, 220, '腾讯云短信应用 SDK AppID', 0, 0, 0, 0, 0),
('sms', '短信配置', '腾讯云短信签名', 'sms_tencent_sign_name', '', 'text', NULL, NULL, 230, '在腾讯云控制台申请的签名', 0, 0, 0, 0, 0)
ON CONFLICT DO NOTHING;
