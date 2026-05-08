-- ============================================================
-- letter_manage_db 初始化脚本
-- 新机器首次部署: mysql -h <host> -u <user> -p < init.sql
-- 注意: Go AutoMigrate 会自动建表，此脚本作为手动备份
-- ============================================================

CREATE DATABASE IF NOT EXISTS letter_manage_db
  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE letter_manage_db;

-- ============================================================
-- DROP（子表→父表顺序，避免外键冲突）
-- ============================================================
DROP TABLE IF EXISTS `letter_special_focuses`;
DROP TABLE IF EXISTS `letter_attachments`;
DROP TABLE IF EXISTS `letter_flows`;
DROP TABLE IF EXISTS `feedbacks`;
DROP TABLE IF EXISTS `letter_signoffs`;
DROP TABLE IF EXISTS `operation_logs`;
DROP TABLE IF EXISTS `dispatch_permissions`;
DROP TABLE IF EXISTS `user_sessions`;
DROP TABLE IF EXISTS `letters`;
DROP TABLE IF EXISTS `police_users`;
DROP TABLE IF EXISTS `special_focuses`;
DROP TABLE IF EXISTS `prompts`;
DROP TABLE IF EXISTS `categories`;
DROP TABLE IF EXISTS `units`;

-- ============================================================
-- CREATE（父表→子表顺序，满足外键依赖）
-- ============================================================

-- 1. 单位表
CREATE TABLE `units` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `level1` varchar(128) DEFAULT NULL,
  `level2` varchar(128) DEFAULT NULL,
  `level3` varchar(128) DEFAULT NULL,
  `system_code` varchar(64) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_units_system_code` (`system_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='单位表';

-- 2. 信件分类表
CREATE TABLE `categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `level1` varchar(64) NOT NULL,
  `level2` varchar(64) DEFAULT NULL,
  `level3` varchar(64) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_level1` (`level1`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='信件分类表';

-- 3. 信件表
CREATE TABLE `letters` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `letter_no` varchar(64) NOT NULL,
  `citizen_name` varchar(64) DEFAULT NULL,
  `phone` varchar(32) DEFAULT NULL,
  `id_card` varchar(32) DEFAULT NULL,
  `received_at` datetime(3) DEFAULT NULL,
  `channel` tinyint DEFAULT NULL COMMENT '1=市民上报,2=局长信箱,7=12337,8=12389,9=进京访,10=12345工单,11=12345专席',
  `category_id` bigint unsigned DEFAULT NULL,
  `content` text,
  `current_status` tinyint DEFAULT NULL COMMENT '1=预处理..10=已办结',
  `current_unit_id` bigint unsigned DEFAULT NULL,
  `handler_user_id` bigint unsigned DEFAULT NULL,
  `handler_unit_id` bigint unsigned DEFAULT NULL,
  `current_operator` varchar(64) DEFAULT NULL,
  `deadline_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_letters_letter_no` (`letter_no`),
  KEY `idx_received_at` (`received_at`),
  KEY `idx_letters_category_id` (`category_id`),
  KEY `fk_letters_current_unit_obj` (`current_unit_id`),
  KEY `idx_handler_user_id` (`handler_user_id`),
  KEY `idx_handler_unit_id` (`handler_unit_id`),
  CONSTRAINT `fk_letters_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`),
  CONSTRAINT `fk_letters_current_unit_obj` FOREIGN KEY (`current_unit_id`) REFERENCES `units` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='信件表';

-- 4. 警察用户表
CREATE TABLE `police_users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `password_hash` varchar(128) NOT NULL,
  `name` varchar(64) NOT NULL,
  `nickname` varchar(64) DEFAULT NULL,
  `police_number` varchar(32) NOT NULL,
  `phone` varchar(32) DEFAULT NULL,
  `permission_level` enum('CITY','DISTRICT','OFFICER') NOT NULL,
  `is_active` tinyint(1) DEFAULT '1',
  `is_admin` tinyint(1) DEFAULT '0',
  `unit_id` bigint unsigned DEFAULT NULL,
  `last_login` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_police_users_police_number` (`police_number`),
  KEY `fk_police_users_unit` (`unit_id`),
  CONSTRAINT `fk_police_users_unit` FOREIGN KEY (`unit_id`) REFERENCES `units` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='警察用户表';

-- 5. 用户会话表
CREATE TABLE `user_sessions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `session_key` varchar(64) NOT NULL,
  `ip_address` varchar(64) DEFAULT NULL,
  `user_agent` varchar(256) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `expires_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_sessions_session_key` (`session_key`),
  KEY `idx_user_sessions_user_id` (`user_id`),
  CONSTRAINT `fk_user_sessions_user` FOREIGN KEY (`user_id`) REFERENCES `police_users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表';

-- 6. 信件流转记录表
CREATE TABLE `letter_flows` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `letter_no` varchar(64) NOT NULL,
  `flow_records` json DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_letter_flows_letter_no` (`letter_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='信件流转记录表';

-- 7. 信件附件表
CREATE TABLE `letter_attachments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `letter_no` varchar(64) NOT NULL,
  `city_dispatch_files` json DEFAULT NULL,
  `district_dispatch_files` json DEFAULT NULL,
  `handler_feedback_files` json DEFAULT NULL,
  `district_feedback_files` json DEFAULT NULL,
  `call_recordings` json DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_letter_attachments_letter_no` (`letter_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='信件附件表';

-- 8. 反馈表
CREATE TABLE `feedbacks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `letter_no` varchar(64) NOT NULL,
  `feedback_info` json DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_feedbacks_letter_no` (`letter_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='反馈表';

-- 9. 下发权限配置表
CREATE TABLE `dispatch_permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `unit_name` varchar(128) NOT NULL,
  `unit_id` bigint unsigned DEFAULT NULL,
  `dispatch_scope` json DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_dispatch_permissions_unit_name` (`unit_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='下发权限配置表';

-- 10. AI提示词表
CREATE TABLE `prompts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `prompt_type` varchar(64) NOT NULL,
  `content` text,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_prompt_type` (`prompt_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI提示词表';

-- 11. 专项关注标签表
CREATE TABLE `special_focuses` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tag_name` varchar(64) NOT NULL,
  `description` text,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_tag_name` (`tag_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='专项关注标签表';

-- 12. 信件-专项关注绑定表
CREATE TABLE `letter_special_focuses` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `letter_no` varchar(64) NOT NULL,
  `focus_id` bigint unsigned NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_lsf_letter_no` (`letter_no`),
  KEY `idx_lsf_focus_id` (`focus_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 13. 操作日志表
CREATE TABLE `operation_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `user_name` varchar(64) DEFAULT '',
  `police_number` varchar(32) DEFAULT '',
  `action` varchar(32) NOT NULL,
  `target` varchar(64) NOT NULL,
  `target_id` varchar(64) DEFAULT '',
  `detail` text,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_operation_logs_user_id` (`user_id`),
  KEY `idx_operation_logs_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 14. 签收办理记录表
CREATE TABLE `letter_signoffs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `letter_no` varchar(64) NOT NULL,
  `action` varchar(32) NOT NULL,
  `from_unit` varchar(256) DEFAULT NULL,
  `to_unit` varchar(256) DEFAULT NULL,
  `operator` varchar(64) DEFAULT NULL,
  `operator_id` bigint unsigned DEFAULT NULL,
  `prev_status` varchar(64) DEFAULT NULL,
  `current_status` varchar(64) DEFAULT NULL,
  `remark` text,
  `recorded_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_signoff_letter_no` (`letter_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='签收办理记录表';


-- ============================================================
-- 种子数据
-- ============================================================

INSERT INTO `categories` (`id`, `level1`, `level2`, `level3`, `created_at`, `updated_at`) VALUES
(1,'投诉举报类','接处警投诉','不出警、出警慢',NOW(),NOW()),
(2,'投诉举报类','执法办案投诉','有案不立',NOW(),NOW()),
(3,'投诉举报类','执法办案投诉','压案不查，久拖不决',NOW(),NOW()),
(4,'投诉举报类','执法办案投诉','追赃挽损效果不明显',NOW(),NOW()),
(5,'投诉举报类','执法办案投诉','交通事故处理速度快慢、不公正',NOW(),NOW()),
(6,'投诉举报类','执法办案投诉','不按法定程序和要求办理案件',NOW(),NOW()),
(7,'投诉举报类','执法办案投诉','违规查扣冻',NOW(),NOW()),
(8,'投诉举报类','执法办案投诉','行政处罚结果不公正',NOW(),NOW()),
(9,'投诉举报类','执法办案投诉','滥用强制措施',NOW(),NOW()),
(10,'投诉举报类','执法办案投诉','插手经济纠纷',NOW(),NOW()),
(11,'投诉举报类','执法办案投诉','乱罚款、乱收费',NOW(),NOW()),
(12,'投诉举报类','执法办案投诉','办人情案、关系案、金钱案',NOW(),NOW()),
(13,'投诉举报类','行政管理服务投诉','推诿扯皮、办事拖拉',NOW(),NOW()),
(14,'投诉举报类','行政管理服务投诉','程序繁琐效率低',NOW(),NOW()),
(15,'投诉举报类','行政管理服务投诉','不一次性告知、让群众反复跑',NOW(),NOW()),
(16,'投诉举报类','行政管理服务投诉','设备设施老化缺失',NOW(),NOW()),
(17,'投诉举报类','队伍纪律作风投诉','吃拿卡要',NOW(),NOW()),
(18,'投诉举报类','队伍纪律作风投诉','对待群众态度粗暴恶劣服务态度差',NOW(),NOW()),
(19,'投诉举报类','队伍纪律作风投诉','耍官威搞特权',NOW(),NOW()),
(20,'投诉举报类','队伍纪律作风投诉','举报民警违纪违法线索',NOW(),NOW()),
(21,'投诉举报类','其他投诉','其他投诉',NOW(),NOW()),
(22,'意见建议类','执法办案建议','案件办理效率',NOW(),NOW()),
(23,'意见建议类','执法办案建议','公正文明执法',NOW(),NOW()),
(24,'意见建议类','执法办案建议','案件反馈/公开',NOW(),NOW()),
(25,'意见建议类','执法办案建议','优化营商环境',NOW(),NOW()),
(26,'意见建议类','政务服务管理建议','车驾管业务方面',NOW(),NOW()),
(27,'意见建议类','政务服务管理建议','户籍业务方面',NOW(),NOW()),
(28,'意见建议类','政务服务管理建议','出入境管理业务方面',NOW(),NOW()),
(29,'意见建议类','政务服务管理建议','优化窗口服务机制模式',NOW(),NOW()),
(30,'意见建议类','政务服务管理建议','完善线上服务模式',NOW(),NOW()),
(31,'意见建议类','政务服务管理建议','政策优化创新方面',NOW(),NOW()),
(32,'意见建议类','社会治理建议','交通秩序方面（停车难、拥堵、乱停放等）',NOW(),NOW()),
(33,'意见建议类','社会治理建议','公众交通与安全出行',NOW(),NOW()),
(34,'意见建议类','社会治理建议','道路设施方面',NOW(),NOW()),
(35,'意见建议类','社会治理建议','社会治安方面（违法犯罪防控）',NOW(),NOW()),
(36,'意见建议类','社会治理建议','公共秩序方面（噪音扰民、遛狗不栓绳等）',NOW(),NOW()),
(37,'意见建议类','社会治理建议','治安行政管理方面',NOW(),NOW()),
(38,'意见建议类','队伍管理建议','加强队伍作风形象建设',NOW(),NOW()),
(39,'意见建议类','队伍管理建议','加强警力和保障',NOW(),NOW()),
(40,'意见建议类','队伍管理建议','提升队伍业务素质',NOW(),NOW()),
(41,'意见建议类','队伍管理建议','加强队伍监督管理',NOW(),NOW()),
(42,'意见建议类','其他意见建议','其他意见建议',NOW(),NOW()),
(43,'咨询政策类','业务咨询','户政业务咨询',NOW(),NOW()),
(44,'咨询政策类','业务咨询','出入境业务咨询',NOW(),NOW()),
(45,'咨询政策类','业务咨询','交管业务咨询',NOW(),NOW()),
(46,'咨询政策类','业务咨询','行政审批业务咨询',NOW(),NOW()),
(47,'咨询政策类','业务咨询','治安业务咨询',NOW(),NOW()),
(48,'咨询政策类','业务咨询','其他业务咨询',NOW(),NOW()),
(49,'咨询政策类','案件咨询','案件办理进度咨询',NOW(),NOW()),
(50,'咨询政策类','案件咨询','追赃挽损情况咨询',NOW(),NOW()),
(51,'咨询政策类','案件咨询','采取强制措施咨询',NOW(),NOW()),
(52,'咨询政策类','案件咨询','其他案件情况咨询',NOW(),NOW()),
(53,'咨询政策类','其他咨询','其他咨询',NOW(),NOW()),
(54,'求助类','案件求助','请求调解案件',NOW(),NOW()),
(55,'求助类','案件求助','请求受理初查',NOW(),NOW()),
(56,'求助类','案件求助','请求加快案件办理',NOW(),NOW()),
(57,'求助类','案件求助','请求加速立案/撤案程序',NOW(),NOW()),
(58,'求助类','案件求助','请求返还涉案财物',NOW(),NOW()),
(59,'求助类','案件求助','请求冻结/解冻银行卡',NOW(),NOW()),
(60,'求助类','案件求助','其他帮助请求',NOW(),NOW()),
(61,'求助类','案件求助','其他事项催办',NOW(),NOW()),
(62,'求助类','非案件求助','请求撤销投诉/举报',NOW(),NOW()),
(63,'求助类','非案件求助','请求维护交通秩序',NOW(),NOW()),
(64,'求助类','非案件求助','请求调解民事纠纷',NOW(),NOW()),
(65,'求助类','非案件求助','请求加强宠物管理',NOW(),NOW()),
(66,'求助类','非案件求助','请求设置/移除/维修交通设施',NOW(),NOW()),
(67,'求助类','非案件求助','请求查处车辆非法改装/炸街/扰民',NOW(),NOW()),
(68,'求助类','非案件求助','请求调取/保存/提供证据',NOW(),NOW()),
(69,'求助类','非案件求助','请求上级协调业务办理/指定管辖',NOW(),NOW()),
(70,'求助类','非案件求助','请求治理烟花爆竹乱燃乱放',NOW(),NOW()),
(71,'求助类','非案件求助','请求帮助寻人/寻物',NOW(),NOW()),
(72,'求助类','其他求助','其他求助',NOW(),NOW()),
(73,'申诉类','执法依据申诉','勤务活动依据申诉',NOW(),NOW()),
(74,'申诉类','执法依据申诉','调查结论依据申诉',NOW(),NOW()),
(75,'申诉类','执法依据申诉','非公安管辖事项依据申诉',NOW(),NOW()),
(76,'申诉类','执法依据申诉','业务不予受理依据申诉',NOW(),NOW()),
(77,'申诉类','行政处罚申诉','申诉已方处罚过重',NOW(),NOW()),
(78,'申诉类','行政处罚申诉','申诉对方处罚过轻',NOW(),NOW()),
(79,'申诉类','行政处罚申诉','申诉遗漏违法行为人',NOW(),NOW()),
(80,'申诉类','行政处罚申诉','申诉增列违法行为人',NOW(),NOW()),
(81,'申诉类','事故认定申诉','申诉己方过错判定',NOW(),NOW()),
(82,'申诉类','事故认定申诉','申诉对方过错判定',NOW(),NOW()),
(83,'申诉类','事故认定申诉','申诉行政行为不当',NOW(),NOW()),
(84,'申诉类','案件申诉','申诉案件定性不当',NOW(),NOW()),
(85,'申诉类','案件申诉','申诉调查取证不当',NOW(),NOW()),
(86,'申诉类','案件申诉','申诉案件办理时限',NOW(),NOW()),
(87,'申诉类','案件申诉','申诉立案/不予立案/撤案',NOW(),NOW()),
(88,'申诉类','其他申诉','其他申诉',NOW(),NOW()),
(89,'表扬肯定类','表扬肯定','整体工作表扬',NOW(),NOW()),
(90,'表扬肯定类','表扬肯定','具体事项感谢',NOW(),NOW()),
(91,'提供社会违法线索类','社会违法线索','涉黄线索',NOW(),NOW()),
(92,'提供社会违法线索类','社会违法线索','涉赌线索',NOW(),NOW()),
(93,'提供社会违法线索类','社会违法线索','涉毒线索',NOW(),NOW()),
(94,'提供社会违法线索类','社会违法线索','涉黑线索',NOW(),NOW()),
(95,'提供社会违法线索类','社会违法线索','涉恐涉政线索',NOW(),NOW()),
(96,'提供社会违法线索类','社会违法线索','交通违法线索',NOW(),NOW()),
(97,'提供社会违法线索类','社会违法线索','其他线索',NOW(),NOW()),
(98,'其他类','不便分类','其他事项',NOW(),NOW());

INSERT INTO `units` (`id`, `level1`, `level2`, `level3`, `system_code`, `created_at`) VALUES
(1,'市局','人事科','','rsk',NOW()),
(2,'市局','宣传科','','xck',NOW()),
(3,'市局','教育训练科','','jyxlk',NOW()),
(9,'市局','情指中心','','qzzx',NOW()),
(10,'市局','科信支队','','kxzd',NOW()),
(11,'市局','信访科','','xfk',NOW()),
(14,'市局','法制支队','','fzzd',NOW()),
(15,'市局','交管直办中心','','jgzbzx',NOW()),
(17,'市局','刑侦支队','','xzzd',NOW()),
(21,'市局','治安支队','','zazd',NOW()),
(24,'市局','经侦支队','','jingzzd',NOW()),
(25,'市局','督审支队','','dszd',NOW()),
(26,'市局','出入境支队','','crjzd',NOW()),
(30,'市局','看守所','','kss',NOW()),
(31,'市局','拘留所','','jls',NOW()),
(32,'分局','冀州区公安局','','jzqgaj',NOW()),
(33,'分局','交管支队','指挥大队','jgzd-996746',NOW()),
(40,'分局','交管支队','桃城大队','jgzd-344443',NOW()),
(41,'分局','交管支队','高新大队','jgzd-723852',NOW()),
(42,'分局','交管支队','滨湖大队','jgzd-893701',NOW()),
(45,'分局','桃城分局','政工室','tcfj-639034',NOW()),
(50,'分局','桃城分局','刑侦一中队','tcfj-370606',NOW()),
(54,'分局','桃城分局','刑侦二中队','tcfj-373975',NOW()),
(55,'分局','桃城分局','治安大队','tcfj-755814',NOW()),
(56,'分局','桃城分局','刑侦三中队','tcfj-691748',NOW()),
(57,'分局','桃城分局','经侦大队','tcfj-537581',NOW()),
(418,'分局','桃城分局','民意智感中心','tcfj-999999',NOW());

-- 默认管理员（密码: 123456，bcrypt hash）
INSERT INTO `police_users` (`id`, `password_hash`, `name`, `nickname`, `police_number`, `phone`, `permission_level`, `is_active`, `is_admin`, `unit_id`, `created_at`) VALUES
(1,'$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy','系统管理员','admin','000001','13303180001','CITY',1,1,418,NOW());

INSERT INTO `prompts` (`id`, `prompt_type`, `content`, `created_at`, `updated_at`) VALUES
(1,'classification','你是一个信件分类助手，请根据以下信件内容判断所属分类。',NOW(),NOW()),
(2,'summary','请对以下信件内容进行摘要总结，字数不超过100字。',NOW(),NOW()),
(3,'keywords','请从以下信件内容中提取关键词，以逗号分隔。',NOW(),NOW());
