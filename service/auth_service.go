package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"letter-manage-backend/dao"
	"letter-manage-backend/model"

	"gorm.io/gorm"
)

func HashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateSessionKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type LoginResult struct {
	SessionKey string
	User       *model.PoliceUser
}

func Login(policeNumber, password string, rememberMe bool, ip, userAgent string) (*LoginResult, error) {
	user, err := dao.GetUserByPoliceNumber(policeNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在或已禁用")
		}
		return nil, err
	}
	if !user.IsActive {
		return nil, errors.New("该用户已被禁用，请联系管理员")
	}
	hashed := HashPassword(password)
	if user.PasswordHash != hashed {
		return nil, errors.New("密码错误")
	}

	sessionKey, err := GenerateSessionKey()
	if err != nil {
		return nil, err
	}

	var expiresAt time.Time
	if rememberMe {
		expiresAt = time.Now().Add(30 * 24 * time.Hour)
	} else {
		expiresAt = time.Now().Add(8 * time.Hour)
	}

	session := &model.UserSession{
		UserID:     user.ID,
		SessionKey: sessionKey,
		IPAddress:  ip,
		UserAgent:  userAgent,
		ExpiresAt:  expiresAt,
	}
	if err := dao.CreateSession(session); err != nil {
		return nil, err
	}
	dao.UpdateUserLastLogin(user.ID)

	return &LoginResult{SessionKey: sessionKey, User: user}, nil
}

func Logout(sessionKey string) error {
	return dao.DeleteSession(sessionKey)
}

func CheckSession(sessionKey string) (*model.PoliceUser, error) {
	session, err := dao.GetSessionByKey(sessionKey)
	if err != nil {
		return nil, err
	}
	return &session.User, nil
}
