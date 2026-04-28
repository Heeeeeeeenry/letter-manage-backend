package dao

import (
	"strings"
	"time"

	"letter-manage-backend/model"

	"gorm.io/gorm"
)

func GetUserByPoliceNumber(policeNumber string) (*model.PoliceUser, error) {
	var user model.PoliceUser
	err := DB.Where("police_number = ? AND is_active = true", policeNumber).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(id uint) (*model.PoliceUser, error) {
	var user model.PoliceUser
	err := DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateSession(session *model.UserSession) error {
	return DB.Create(session).Error
}

func GetSessionByKey(sessionKey string) (*model.UserSession, error) {
	var session model.UserSession
	err := DB.Preload("User").Where("session_key = ? AND expires_at > ?", sessionKey, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func DeleteSession(sessionKey string) error {
	return DB.Where("session_key = ?", sessionKey).Delete(&model.UserSession{}).Error
}

func DeleteUserSessions(userID uint) error {
	return DB.Where("user_id = ?", userID).Delete(&model.UserSession{}).Error
}

func UpdateUserLastLogin(userID uint) error {
	now := time.Now()
	return DB.Model(&model.PoliceUser{}).Where("id = ?", userID).Update("last_login", now).Error
}

func CreateUser(user *model.PoliceUser) error {
	return DB.Create(user).Error
}

func UpdateUser(user *model.PoliceUser) error {
	return DB.Save(user).Error
}

func DeleteUser(id uint) error {
	return DB.Delete(&model.PoliceUser{}, id).Error
}

func GetUserList(page, pageSize int, unitFilter string, permLevel string) ([]model.PoliceUser, int64, error) {
	var users []model.PoliceUser
	var total int64
	offset := (page - 1) * pageSize
	query := DB.Model(&model.PoliceUser{})
	if unitFilter != "" && permLevel == "DISTRICT" {
		// 区县局：看到本单位本级用户 + 下属科所队用户，但排除 CITY 级别用户
		subUnits := GetSubordinateUnitNames(unitFilter)
		if len(subUnits) > 0 {
			query = query.Where(
				"(unit_name = ?) OR (unit_name IN ? AND permission_level != 'CITY')",
				unitFilter, subUnits,
			)
		} else {
			query = query.Where("unit_name = ?", unitFilter)
		}
	} else if unitFilter != "" {
		query = query.Where("unit_name = ?", unitFilter)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func GetAllUsers() ([]model.PoliceUser, error) {
	var users []model.PoliceUser
	err := DB.Find(&users).Error
	return users, err
}

func CleanExpiredSessions() error {
	return DB.Where("expires_at < ?", time.Now()).Delete(&model.UserSession{}).Error
}

// Unit DAO

func GetUnitList(page, pageSize int) ([]model.Unit, int64, error) {
	var units []model.Unit
	var total int64
	offset := (page - 1) * pageSize
	query := DB.Model(&model.Unit{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Offset(offset).Limit(pageSize).Find(&units).Error
	return units, total, err
}

func GetAllUnits() ([]model.Unit, error) {
	var units []model.Unit
	err := DB.Find(&units).Error
	return units, err
}

// GetUnitsWithFilter 获取单位列表，支持分页和筛选
func GetUnitsWithFilter(page, pageSize int, searchKeyword, filterLevel1, filterLevel2 string) ([]model.Unit, int64, error) {
	var units []model.Unit
	var total int64

	query := DB.Model(&model.Unit{})

	// 搜索关键词：匹配 level1, level2, level3 任一字段
	if searchKeyword != "" {
		keyword := "%" + searchKeyword + "%"
		query = query.Where("level1 LIKE ? OR level2 LIKE ? OR level3 LIKE ?", keyword, keyword, keyword)
	}

	// 一级单位筛选
	if filterLevel1 != "" {
		query = query.Where("level1 = ?", filterLevel1)
	}

	// 二级单位筛选
	if filterLevel2 != "" {
		query = query.Where("level2 = ?", filterLevel2)
	}

	// 计算总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页：如果 pageSize <= 0，则返回所有数据
	if pageSize > 0 {
		offset := (page - 1) * pageSize
		err = query.Offset(offset).Limit(pageSize).Find(&units).Error
	} else {
		// 返回所有数据
		err = query.Find(&units).Error
	}
	if err != nil {
		return nil, 0, err
	}

	return units, total, err
}
// normalizeUnitName 将全路径单位名转为短名
// "分局 / 桃城分局 / 民意智感中心" → "民意智感中心"
// "民意智感中心" → "民意智感中心"（不变）
func normalizeUnitName(name string) string {
	parts := strings.Split(name, " / ")
	return strings.TrimSpace(parts[len(parts)-1])
}

// GetSubordinateUnitNames 获取某单位及其下属所有单位的短名称列表
// unitName 支持全路径格式（如"分局 / 桃城分局 / 民意智感中心"）和短名格式
func GetSubordinateUnitNames(unitName string) []string {
	allUnits, err := GetAllUnits()
	if err != nil {
		return nil
	}
	// 归一化：提取最后一段作为匹配依据
	shortName := normalizeUnitName(unitName)
	var names []string
	seen := map[string]bool{}
	for _, u := range allUnits {
		if u.Level1 == shortName || u.Level2 == shortName || u.Level3 == shortName {
			shortName := u.Level3
			if shortName == "" {
				shortName = u.Level2
			}
			if shortName == "" {
				shortName = u.Level1
			}
			if shortName != "" && !seen[shortName] {
				seen[shortName] = true
				names = append(names, shortName)
			}
		}
	}
	return names
}

func GetUnitByID(id uint) (*model.Unit, error) {
	var unit model.Unit
	err := DB.First(&unit, id).Error
	if err != nil {
		return nil, err
	}
	return &unit, nil
}

func CreateUnit(unit *model.Unit) error {
	return DB.Create(unit).Error
}

func UpdateUnit(unit *model.Unit) error {
	return DB.Save(unit).Error
}

func DeleteUnit(id uint) error {
	return DB.Delete(&model.Unit{}, id).Error
}

// DispatchPermission DAO

func GetDispatchPermissions() ([]model.DispatchPermission, error) {
	var perms []model.DispatchPermission
	err := DB.Find(&perms).Error
	return perms, err
}

func GetDispatchPermissionByUnit(unitName string) (*model.DispatchPermission, error) {
	var perm model.DispatchPermission
	err := DB.Where("unit_name = ?", unitName).First(&perm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &perm, nil
}

func GetDispatchPermissionByID(id uint) (*model.DispatchPermission, error) {
	var perm model.DispatchPermission
	err := DB.First(&perm, id).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func CreateDispatchPermission(perm *model.DispatchPermission) error {
	return DB.Create(perm).Error
}

func UpdateDispatchPermission(perm *model.DispatchPermission) error {
	return DB.Save(perm).Error
}

func DeleteDispatchPermission(id uint) error {
	return DB.Delete(&model.DispatchPermission{}, id).Error
}
