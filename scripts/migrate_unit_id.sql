-- =====================================================
-- 数据迁移脚本：为 police_users, letters, dispatch_permissions
-- 增加 unit_id / current_unit_id 外键列，并回填数据
-- =====================================================

-- 1. police_users 表
ALTER TABLE police_users
  ADD COLUMN unit_id INT NULL AFTER phone,
  ADD INDEX idx_unit_id (unit_id);

-- 回填 unit_id：通过 unit_name 匹配 units 表
UPDATE police_users pu
  JOIN units u ON (
    pu.unit_name = u.level3
    OR pu.unit_name = u.level2
    OR pu.unit_name = u.level1
  )
SET pu.unit_id = u.id
WHERE pu.unit_id IS NULL;

-- 对于全路径格式的 unit_name（包含 " / "），尝试用最后一段匹配
UPDATE police_users pu
  JOIN units u ON (
    SUBSTRING_INDEX(pu.unit_name, ' / ', -1) = u.level3
    OR SUBSTRING_INDEX(pu.unit_name, ' / ', -1) = u.level2
    OR SUBSTRING_INDEX(pu.unit_name, ' / ', -1) = u.level1
  )
SET pu.unit_id = u.id
WHERE pu.unit_id IS NULL;

-- 2. letters 表
ALTER TABLE letters
  ADD COLUMN current_unit_id INT NULL AFTER special_tags,
  ADD INDEX idx_current_unit_id (current_unit_id);

-- 回填 current_unit_id：通过 current_unit 匹配 units 表
UPDATE letters l
  JOIN units u ON (
    l.current_unit = u.level3
    OR l.current_unit = u.level2
    OR l.current_unit = u.level1
  )
SET l.current_unit_id = u.id
WHERE l.current_unit_id IS NULL;

-- 全路径格式回填
UPDATE letters l
  JOIN units u ON (
    SUBSTRING_INDEX(l.current_unit, ' / ', -1) = u.level3
    OR SUBSTRING_INDEX(l.current_unit, ' / ', -1) = u.level2
    OR SUBSTRING_INDEX(l.current_unit, ' / ', -1) = u.level1
  )
SET l.current_unit_id = u.id
WHERE l.current_unit_id IS NULL;

-- 3. dispatch_permissions 表
ALTER TABLE dispatch_permissions
  ADD COLUMN unit_id INT NULL AFTER id,
  ADD INDEX idx_dp_unit_id (unit_id);

-- 回填 unit_id：通过 unit_name 匹配 units 表
UPDATE dispatch_permissions dp
  JOIN units u ON (
    dp.unit_name = u.level3
    OR dp.unit_name = u.level2
    OR dp.unit_name = u.level1
  )
SET dp.unit_id = u.id
WHERE dp.unit_id IS NULL;

-- 查看回填结果
SELECT 'police_users' AS table_name, COUNT(*) AS total, SUM(CASE WHEN unit_id IS NOT NULL THEN 1 ELSE 0 END) AS filled FROM police_users
UNION ALL
SELECT 'letters', COUNT(*), SUM(CASE WHEN current_unit_id IS NOT NULL THEN 1 ELSE 0 END) FROM letters
UNION ALL
SELECT 'dispatch_permissions', COUNT(*), SUM(CASE WHEN unit_id IS NOT NULL THEN 1 ELSE 0 END) FROM dispatch_permissions;
