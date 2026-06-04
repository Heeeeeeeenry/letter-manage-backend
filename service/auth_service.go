package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"letter-manage-backend/dao"
	"letter-manage-backend/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// fallback to SHA256 for safety
		h := sha256.New()
		h.Write([]byte(password))
		return hex.EncodeToString(h.Sum(nil))
	}
	return string(hash)
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
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
	// Try bcrypt first, fallback to SHA256 for legacy passwords
	if !CheckPassword(password, user.PasswordHash) {
		// legacy SHA256 fallback
		h := sha256.New()
		h.Write([]byte(password))
		legacyHash := hex.EncodeToString(h.Sum(nil))
		if user.PasswordHash != legacyHash {
			return nil, errors.New("密码错误")
		}
		// Auto-upgrade legacy hash to bcrypt
		newHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		dao.UpdateUserPassword(user.ID, string(newHash))
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
