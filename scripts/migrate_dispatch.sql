-- Migration: dispatch_permissions → dispatch_targets
-- 1. Create new table
CREATE TABLE IF NOT EXISTS `dispatch_targets` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `dispatcher_unit_id` bigint unsigned NOT NULL COMMENT '下发发起单位ID',
  `target_unit_id` bigint unsigned NOT NULL COMMENT '可下发目标单位ID',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_dispatcher_target` (`dispatcher_unit_id`, `target_unit_id`),
  KEY `idx_dispatcher` (`dispatcher_unit_id`),
  KEY `idx_target` (`target_unit_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='下发权限关系表';

-- 2. Migrate data: parse dispatch_permissions JSON scope → dispatch_targets rows
-- For each dispatch_permission, parse dispatch_scope JSON array, resolve unit names→IDs, insert
-- Handled by Go migration code (see migrate_dispatch_permissions.go)

-- 3. Drop old table (after migration verified)
-- DROP TABLE IF EXISTS dispatch_permissions;
