package dao

import (
	"encoding/json"
	"time"

	"letter-manage-backend/model"

	"gorm.io/gorm"
)

type LetterFilter struct {
	Status     string
	CategoryL1 string
	CategoryL2 string
	CategoryL3 string
	Keyword    string
	LetterNo   string
	CitizenName string
	Phone      string
	IDCard     string
	StartTime  *time.Time
	EndTime    *time.Time
	UnitName   string
	UnitNames  []string
	Page       int
	PageSize   int
}

func buildLetterQuery(filter LetterFilter) *gorm.DB {
	query := DB.Model(&model.Letter{})
	if filter.Status != "" {
		query = query.Where("current_status = ?", filter.Status)
	}
	if filter.CategoryL1 != "" {
		query = query.Where("category_l1 = ?", filter.CategoryL1)
	}
	if filter.CategoryL2 != "" {
		query = query.Where("category_l2 = ?", filter.CategoryL2)
	}
	if filter.CategoryL3 != "" {
		query = query.Where("category_l3 = ?", filter.CategoryL3)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where("citizen_name LIKE ? OR phone LIKE ? OR letter_no LIKE ? OR content LIKE ?", like, like, like, like)
	}
	if filter.LetterNo != "" {
		query = query.Where("letter_no = ?", filter.LetterNo)
	}
	if filter.CitizenName != "" {
		query = query.Where("citizen_name = ?", filter.CitizenName)
	}
	if filter.Phone != "" {
		query = query.Where("phone = ?", filter.Phone)
	}
	if filter.IDCard != "" {
		query = query.Where("id_card = ?", filter.IDCard)
	}
	if filter.StartTime != nil {
		query = query.Where("received_at >= ?", filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("received_at <= ?", filter.EndTime)
	}
	if filter.UnitName != "" {
		query = query.Where("current_unit = ?", filter.UnitName)
	}
	if len(filter.UnitNames) > 0 {
		query = query.Where("current_unit IN ?", filter.UnitNames)
	}
	return query
}

func GetLetterList(filter LetterFilter) ([]model.Letter, int64, error) {
	var letters []model.Letter
	var total int64
	query := buildLetterQuery(filter)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&letters).Error
	return letters, total, err
}

func GetLetterByNo(letterNo string) (*model.Letter, error) {
	var letter model.Letter
	err := DB.Where("letter_no = ?", letterNo).First(&letter).Error
	if err != nil {
		return nil, err
	}
	return &letter, nil
}

func GetLetterByID(id uint) (*model.Letter, error) {
	var letter model.Letter
	err := DB.First(&letter, id).Error
	if err != nil {
		return nil, err
	}
	return &letter, nil
}

func GetLettersByPhone(phone string) ([]model.Letter, error) {
	var letters []model.Letter
	err := DB.Where("phone = ?", phone).Order("created_at DESC").Find(&letters).Error
	return letters, err
}

func GetLettersByIDCard(idCard string) ([]model.Letter, error) {
	var letters []model.Letter
	err := DB.Where("id_card = ?", idCard).Order("created_at DESC").Find(&letters).Error
	return letters, err
}

func CreateLetter(letter *model.Letter) error {
	return DB.Create(letter).Error
}

func UpdateLetter(letter *model.Letter) error {
	return DB.Save(letter).Error
}

func DeleteLetter(id uint) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var letter model.Letter
		if err := tx.First(&letter, id).Error; err != nil {
			return err
		}
		if err := tx.Where("letter_no = ?", letter.LetterNo).Delete(&model.LetterFlow{}).Error; err != nil {
			return err
		}
		if err := tx.Where("letter_no = ?", letter.LetterNo).Delete(&model.LetterAttachment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("letter_no = ?", letter.LetterNo).Delete(&model.Feedback{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Letter{}, id).Error
	})
}

func UpdateLetterOperator(letterNo, operator string) error {
	updates := map[string]interface{}{
		"current_operator": operator,
	}
	return DB.Model(&model.Letter{}).Where("letter_no = ?", letterNo).Updates(updates).Error
}

func UpdateLetterStatus(letterNo, status, unitName string) error {
	updates := map[string]interface{}{
		"current_status": status,
	}
	if unitName != "" {
		updates["current_unit"] = unitName
	}
	return DB.Model(&model.Letter{}).Where("letter_no = ?", letterNo).Updates(updates).Error
}

func UpdateLetterDeadline(letterNo string, deadline *time.Time) error {
	updates := map[string]interface{}{
		"deadline_at": deadline,
	}
	return DB.Model(&model.Letter{}).Where("letter_no = ?", letterNo).Updates(updates).Error
}

// UpdateLetterFields 批量更新信件字段
func UpdateLetterFields(letterNo string, fields map[string]interface{}) error {
	return DB.Model(&model.Letter{}).Where("letter_no = ?", letterNo).Updates(fields).Error
}

// LetterFlow DAO

func GetFlowByLetterNo(letterNo string) (*model.LetterFlow, error) {
	var flow model.LetterFlow
	err := DB.Where("letter_no = ?", letterNo).First(&flow).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &flow, nil
}

func UpsertLetterFlow(letterNo string, flowRecords json.RawMessage) error {
	var flow model.LetterFlow
	err := DB.Where("letter_no = ?", letterNo).First(&flow).Error
	if err == gorm.ErrRecordNotFound {
		flow = model.LetterFlow{
			LetterNo:    letterNo,
			FlowRecords: model.JSONRaw(flowRecords),
		}
		return DB.Create(&flow).Error
	} else if err != nil {
		return err
	}
	flow.FlowRecords = model.JSONRaw(flowRecords)
	return DB.Save(&flow).Error
}

// Attachment DAO

func GetAttachmentByLetterNo(letterNo string) (*model.LetterAttachment, error) {
	var att model.LetterAttachment
	err := DB.Where("letter_no = ?", letterNo).First(&att).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &att, nil
}

func UpsertAttachment(att *model.LetterAttachment) error {
	var existing model.LetterAttachment
	err := DB.Where("letter_no = ?", att.LetterNo).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return DB.Create(att).Error
	} else if err != nil {
		return err
	}
	att.ID = existing.ID
	att.CreatedAt = existing.CreatedAt
	return DB.Save(att).Error
}

// Feedback DAO

func GetFeedbacksByLetterNo(letterNo string) ([]model.Feedback, error) {
	var feedbacks []model.Feedback
	err := DB.Where("letter_no = ?", letterNo).Order("created_at ASC").Find(&feedbacks).Error
	return feedbacks, err
}

func CreateFeedback(fb *model.Feedback) error {
	return DB.Create(fb).Error
}

// Statistics

type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

func GetLetterStatusStats(startTime, endTime *time.Time) ([]StatusCount, error) {
	var results []StatusCount
	query := DB.Model(&model.Letter{}).
		Select("current_status as status, count(*) as count")
	if startTime != nil {
		query = query.Where("received_at >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("received_at <= ?", endTime)
	}
	err := query.Group("current_status").
		Scan(&results).Error
	return results, err
}

type ChannelCount struct {
	Channel string `json:"channel"`
	Count   int64  `json:"count"`
}

func GetLetterChannelStats() ([]ChannelCount, error) {
	var results []ChannelCount
	err := DB.Model(&model.Letter{}).
		Select("channel, count(*) as count").
		Group("channel").
		Scan(&results).Error
	return results, err
}

type MonthCount struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}

func GetLetterMonthStats() ([]MonthCount, error) {
	var results []MonthCount
	err := DB.Model(&model.Letter{}).
		Select("DATE_FORMAT(received_at, '%Y-%m') as month, count(*) as count").
		Group("month").
		Order("month ASC").
		Limit(12).
		Scan(&results).Error
	return results, err
}

func GetDispatchList(unitName string, permLevel string, page, pageSize int) ([]model.Letter, int64, error) {
	query := DB.Model(&model.Letter{})
	switch permLevel {
	case "CITY":
		query = query.Where("current_status = ?", model.StatusPreProcess)
	case "DISTRICT":
		query = query.Where("current_status = ? AND current_unit = ?", model.StatusCityDispatched, unitName)
	default:
		query = query.Where("1 = 0")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	var letters []model.Letter
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&letters).Error
	return letters, total, err
}

func GetProcessingList(unitName string, permLevel string, page, pageSize int) ([]model.Letter, int64, error) {
	query := DB.Model(&model.Letter{})

	// 市局可看「预处理」（未下发）和本单位已下发的信件
	// 区县局/基层单位只可看本单位已下发的信件
	if permLevel == "CITY" {
		query = query.Where(
			"(current_status = ?) OR (current_unit = ? AND current_status IN ?)",
			model.StatusPreProcess,
			unitName,
			[]string{model.StatusDispatched, model.StatusProcessing, model.StatusCityDirectDispatch},
		)
	} else {
		query = query.Where(
			"current_unit = ? AND current_status IN ?",
			unitName,
			[]string{model.StatusDispatched, model.StatusProcessing, model.StatusCityDirectDispatch},
		)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	var letters []model.Letter
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&letters).Error
	return letters, total, err
}

func GetAuditList(unitName string, permLevel string, page, pageSize int) ([]model.Letter, int64, error) {
	query := DB.Model(&model.Letter{})
	switch permLevel {
	case "CITY":
		// 市局：查看分县局已审核的信件 + 越级下发后基层已反馈的信件
		query = query.Where("current_status IN ?", []string{model.StatusPendingCityAudit, model.StatusPendingDistrictAudit})
	case "DISTRICT":
		// 分县局：查看本单位科室已反馈、待分县局审核的信件
		query = query.Where("current_status = ? AND current_unit = ?", model.StatusPendingDistrictAudit, unitName)
	default:
		query = query.Where("1 = 0")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	var letters []model.Letter
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&letters).Error
	return letters, total, err
}
