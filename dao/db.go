package dao

import (
	"fmt"
	"time"

	"letter-manage-backend/config"
	"letter-manage-backend/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {
	cfg := config.Get().Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.Charset)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	return nil
}

func AutoMigrate() error {
	return DB.AutoMigrate(
		&model.Letter{},
		&model.LetterFlow{},
		&model.LetterAttachment{},
		&model.Feedback{},
		&model.Category{},
		&model.PoliceUser{},
		&model.UserSession{},
		&model.Unit{},
		&model.DispatchPermission{},
		&model.Prompt{},
		&model.SpecialFocus{},
	)
}
