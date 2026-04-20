package dao

import (
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

func GetUserList(page, pageSize int) ([]model.PoliceUser, int64, error) {
	var users []model.PoliceUser
	var total int64
	offset := (page - 1) * pageSize
	query := DB.Model(&model.PoliceUser{})
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
