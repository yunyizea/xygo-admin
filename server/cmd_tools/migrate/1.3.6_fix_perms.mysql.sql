-- ============================================================
-- v1.2.6 修正 admin_menu.perms 数据 —— PostgreSQL 版
-- A. 修正 21 个格式不统一的 perms 为标准 JSON 数组
-- B. 补齐 35 个空 perms
-- ============================================================

-- ========== A. 修正格式不统一的 perms ==========

-- 附件管理（冒号格式 -> JSON 数组）
UPDATE xy_admin_menu SET perms = '["GET /admin/attachment/list"]' WHERE id = 123;
UPDATE xy_admin_menu SET perms = '["POST /admin/attachment/edit"]' WHERE id = 124;
UPDATE xy_admin_menu SET perms = '["POST /admin/attachment/delete"]' WHERE id = 125;

-- 定时任务（裸路径 -> JSON 数组）
UPDATE xy_admin_menu SET perms = '["GET /admin/cron/list"]' WHERE id = 241;
UPDATE xy_admin_menu SET perms = '["POST /admin/cron/save"]' WHERE id = 242;
UPDATE xy_admin_menu SET perms = '["POST /admin/cron/delete"]' WHERE id = 243;
UPDATE xy_admin_menu SET perms = '["POST /admin/cron/onlineExec"]' WHERE id = 244;

-- CMS 文档分类
UPDATE xy_admin_menu SET perms = '["POST /admin/cms/docCategory/save"]' WHERE id = 634;
UPDATE xy_admin_menu SET perms = '["POST /admin/cms/docCategory/delete"]' WHERE id = 635;

-- CMS 文档
UPDATE xy_admin_menu SET perms = '["POST /admin/cms/doc/save"]' WHERE id = 632;
UPDATE xy_admin_menu SET perms = '["POST /admin/cms/doc/delete"]' WHERE id = 633;

-- CMS 案例
UPDATE xy_admin_menu SET perms = '["POST /admin/cms/case/save"]' WHERE id = 642;
UPDATE xy_admin_menu SET perms = '["POST /admin/cms/case/delete"]' WHERE id = 643;
UPDATE xy_admin_menu SET perms = '["POST /admin/cms/caseCategory/save"]' WHERE id = 644;
UPDATE xy_admin_menu SET perms = '["POST /admin/cms/caseCategory/delete"]' WHERE id = 645;

-- CMS 社区
UPDATE xy_admin_menu SET perms = '["POST /admin/cms/community/postUpdate"]' WHERE id = 651;
UPDATE xy_admin_menu SET perms = '["POST /admin/cms/community/postDelete"]' WHERE id = 652;
UPDATE xy_admin_menu SET perms = '["POST /admin/cms/community/categorySave"]' WHERE id = 653;
UPDATE xy_admin_menu SET perms = '["POST /admin/cms/community/replyDelete"]' WHERE id = 654;

-- CMS 扩展包
UPDATE xy_admin_menu SET perms = '["POST /admin/addon/save"]' WHERE id = 661;
UPDATE xy_admin_menu SET perms = '["POST /admin/addon/delete"]' WHERE id = 662;


-- ========== B. 补齐空 perms ==========

-- 后台用户 (parent_id=61)
UPDATE xy_admin_menu SET perms = '["POST /admin/user/save"]' WHERE id = 777;
UPDATE xy_admin_menu SET perms = '["POST /admin/user/save","GET /admin/user/detail"]' WHERE id = 778;
UPDATE xy_admin_menu SET perms = '["POST /admin/user/delete"]' WHERE id = 779;
UPDATE xy_admin_menu SET perms = '["POST /admin/user/kick"]' WHERE id = 780;

-- 角色管理 (parent_id=62)
UPDATE xy_admin_menu SET perms = '["POST /admin/role/save"]' WHERE id = 781;
UPDATE xy_admin_menu SET perms = '["POST /admin/role/save","GET /admin/role/detail"]' WHERE id = 782;
UPDATE xy_admin_menu SET perms = '["POST /admin/role/delete"]' WHERE id = 783;
UPDATE xy_admin_menu SET perms = '["POST /admin/role/bindMenus","GET /admin/role/menuIds"]' WHERE id = 784;
UPDATE xy_admin_menu SET perms = '["POST /admin/role/dataScopeEdit"]' WHERE id = 785;
UPDATE xy_admin_menu SET perms = '["GET /admin/fieldPerm/list","POST /admin/fieldPerm/batchSave"]' WHERE id = 786;

-- 附件管理 补充 (parent_id=122)
UPDATE xy_admin_menu SET perms = '["POST /admin/upload/file"]' WHERE id = 793;
UPDATE xy_admin_menu SET perms = '["POST /admin/attachment/delete"]' WHERE id = 794;

-- 部门管理 (parent_id=141)
UPDATE xy_admin_menu SET perms = '["POST /admin/dept/save"]' WHERE id = 787;
UPDATE xy_admin_menu SET perms = '["POST /admin/dept/save","GET /admin/dept/detail"]' WHERE id = 788;
UPDATE xy_admin_menu SET perms = '["POST /admin/dept/delete"]' WHERE id = 789;

-- 岗位管理 (parent_id=142)
UPDATE xy_admin_menu SET perms = '["POST /admin/post/save"]' WHERE id = 790;
UPDATE xy_admin_menu SET perms = '["POST /admin/post/save","GET /admin/post/detail"]' WHERE id = 791;
UPDATE xy_admin_menu SET perms = '["POST /admin/post/delete"]' WHERE id = 792;

-- 会员列表 (parent_id=144)
UPDATE xy_admin_menu SET perms = '["POST /admin/member/add"]' WHERE id = 145;
UPDATE xy_admin_menu SET perms = '["PUT /admin/member/edit","GET /admin/member/detail"]' WHERE id = 146;
UPDATE xy_admin_menu SET perms = '["DELETE /admin/member/delete"]' WHERE id = 147;
UPDATE xy_admin_menu SET perms = '["PUT /admin/member/resetPassword"]' WHERE id = 148;
UPDATE xy_admin_menu SET perms = '["DELETE /admin/member/delete"]' WHERE id = 797;

-- 通知管理 补充 add (parent_id=220)
UPDATE xy_admin_menu SET perms = '["POST /admin/notice/edit"]' WHERE id = 795;

-- 定时任务 补充 add (parent_id=240)
UPDATE xy_admin_menu SET perms = '["POST /admin/cron/save"]' WHERE id = 796;

-- 操作日志 补充 batchDel (parent_id=160)
UPDATE xy_admin_menu SET perms = '["POST /admin/log/operation/delete"]' WHERE id = 798;

-- 登录日志 补充 batchDel (parent_id=157)
UPDATE xy_admin_menu SET perms = '["POST /admin/log/login/delete"]' WHERE id = 799;

-- 余额变动日志 补充 batchDel (parent_id=511)
UPDATE xy_admin_menu SET perms = '["POST /admin/member-money-log/delete"]' WHERE id = 806;

-- 积分变动日志 (parent_id=517) — 全部缺失
UPDATE xy_admin_menu SET perms = '["POST /admin/member-score-log/edit"]' WHERE id = 800;
UPDATE xy_admin_menu SET perms = '["GET /admin/member-score-log/view"]' WHERE id = 801;
UPDATE xy_admin_menu SET perms = '["POST /admin/member-score-log/edit","GET /admin/member-score-log/view"]' WHERE id = 802;
UPDATE xy_admin_menu SET perms = '["POST /admin/member-score-log/delete"]' WHERE id = 803;
UPDATE xy_admin_menu SET perms = '["POST /admin/member-score-log/delete"]' WHERE id = 804;
UPDATE xy_admin_menu SET perms = '["GET /admin/member-score-log/export"]' WHERE id = 805;

-- 会员通知 补充 batchDel (parent_id=617)
UPDATE xy_admin_menu SET perms = '["POST /admin/member-notice/delete"]' WHERE id = 807;
