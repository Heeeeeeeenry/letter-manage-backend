-- =====================================================
-- letter_manage_db DDL
-- Database: letter_manage_db
-- Charset: utf8mb4
-- =====================================================

CREATE DATABASE IF NOT EXISTS letter_manage_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE letter_manage_db;

-- 信件表 letters
CREATE TABLE IF NOT EXISTS `letters` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `letter_no` varchar(64) NOT NULL COMMENT '信件编号 XJ+时间戳',
  `citizen_name` varchar(64) DEFAULT '' COMMENT '群众姓名',
  `phone` varchar(32) DEFAULT '' COMMENT '手机号',
  `id_card` varchar(32) DEFAULT '' COMMENT '身份证号',
  `received_at` datetime NOT NULL COMMENT '来信时间',
  `channel` varchar(64) DEFAULT '' COMMENT '来信渠道',
  `category_l1` varchar(64) DEFAULT '' COMMENT '信件一级分类',
  `category_l2` varchar(64) DEFAULT '' COMMENT '信件二级分类',
  `category_l3` varchar(64) DEFAULT '' COMMENT '信件三级分类',
  `content` text COMMENT '诉求内容',
  `special_tags` json DEFAULT NULL COMMENT '专项关注标签(JSON数组)',
  `current_unit` varchar(128) DEFAULT '' COMMENT '当前信件处理单位',
  `current_status` varchar(64) DEFAULT '预处理' COMMENT '当前信件状态',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_letter_no` (`letter_no`),
  KEY `idx_status` (`current_status`),
  KEY `idx_phone` (`phone`),
  KEY `idx_id_card` (`id_card`),
  KEY `idx_current_unit` (`current_unit`),
  KEY `idx_received_at` (`received_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='信件表';

-- 流转记录表 letter_flows
CREATE TABLE IF NOT EXISTS `letter_flows` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `letter_no` varchar(64) NOT NULL COMMENT '信件编号',
  `flow_records` json DEFAULT NULL COMMENT '流转记录(JSON数组)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_letter_no` (`letter_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='信件流转记录表';

-- 附件表 letter_attachments
CREATE TABLE IF NOT EXISTS `letter_attachments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `letter_no` varchar(64) NOT NULL COMMENT '信件编号',
  `city_dispatch_files` json DEFAULT NULL COMMENT '市局下发附件(JSON数组)',
  `district_dispatch_files` json DEFAULT NULL COMMENT '区县局下发附件(JSON数组)',
  `handler_feedback_files` json DEFAULT NULL COMMENT '办案单位反馈附件(JSON数组)',
  `district_feedback_files` json DEFAULT NULL COMMENT '区县局反馈附件(JSON数组)',
  `call_recordings` json DEFAULT NULL COMMENT '通话录音附件(JSON数组)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_letter_no` (`letter_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='信件附件表';

-- 反馈表 feedbacks
CREATE TABLE IF NOT EXISTS `feedbacks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `letter_no` varchar(64) NOT NULL COMMENT '信件编号',
  `feedback_info` json DEFAULT NULL COMMENT '反馈信息(JSON)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_letter_no` (`letter_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='反馈表';

-- 信件分类表 categories
CREATE TABLE IF NOT EXISTS `categories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `level1` varchar(64) NOT NULL COMMENT '一级分类',
  `level2` varchar(64) DEFAULT '' COMMENT '二级分类',
  `level3` varchar(64) DEFAULT '' COMMENT '三级分类',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_level1` (`level1`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='信件分类表';

-- 单位表 units
CREATE TABLE IF NOT EXISTS `units` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `level1` varchar(128) DEFAULT '' COMMENT '一级单位',
  `level2` varchar(128) DEFAULT '' COMMENT '二级单位',
  `level3` varchar(128) DEFAULT '' COMMENT '三级单位',
  `system_code` varchar(64) NOT NULL COMMENT '系统编码(唯一)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_system_code` (`system_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='单位表';

-- 用户表 police_users
CREATE TABLE IF NOT EXISTS `police_users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `password_hash` varchar(128) NOT NULL COMMENT '密码SHA256哈希',
  `name` varchar(64) NOT NULL COMMENT '真实姓名',
  `nickname` varchar(64) DEFAULT '' COMMENT '昵称',
  `police_number` varchar(32) NOT NULL COMMENT '警号(唯一)',
  `phone` varchar(32) DEFAULT '' COMMENT '手机号',
  `unit_name` varchar(128) DEFAULT '' COMMENT '所属单位',
  `permission_level` enum('CITY','DISTRICT','OFFICER') NOT NULL DEFAULT 'OFFICER' COMMENT '权限级别',
  `is_active` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否启用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `last_login` datetime DEFAULT NULL COMMENT '最后登录时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_police_number` (`police_number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='警察用户表';

-- 会话表 user_sessions
CREATE TABLE IF NOT EXISTS `user_sessions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `session_key` varchar(64) NOT NULL COMMENT '会话密钥(64字符hex)',
  `ip_address` varchar(64) DEFAULT '' COMMENT 'IP地址',
  `user_agent` varchar(256) DEFAULT '' COMMENT 'User-Agent',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `expires_at` datetime NOT NULL COMMENT '过期时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_session_key` (`session_key`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_expires_at` (`expires_at`),
  CONSTRAINT `fk_session_user` FOREIGN KEY (`user_id`) REFERENCES `police_users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表';

-- 下发权限表 dispatch_permissions
CREATE TABLE IF NOT EXISTS `dispatch_permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `unit_name` varchar(128) NOT NULL COMMENT '单位全称',
  `dispatch_scope` json DEFAULT NULL COMMENT '下发范围(JSON数组，单位名列表)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_unit_name` (`unit_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='下发权限配置表';

-- 提示词表 prompts
CREATE TABLE IF NOT EXISTS `prompts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `prompt_type` varchar(64) NOT NULL COMMENT '提示词类型',
  `content` text COMMENT '提示词内容',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_prompt_type` (`prompt_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI提示词表';

-- 专项关注表 special_focuses
CREATE TABLE IF NOT EXISTS `special_focuses` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `tag_name` varchar(64) NOT NULL COMMENT '标签名',
  `description` text COMMENT '描述',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_tag_name` (`tag_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='专项关注标签表';

-- =====================================================
-- 初始数据
-- =====================================================

-- 默认管理员账号 (密码 admin123 的SHA256: 0192023a7bbd73250516f069df18b500)
-- SHA256('admin123') = 240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9
INSERT IGNORE INTO `police_users` (`name`, `nickname`, `police_number`, `password_hash`, `unit_name`, `permission_level`, `is_active`)
VALUES ('系统管理员', 'admin', 'admin', '240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9', '市局', 'CITY', 1);
