package dao

import (
	"encoding/json"
	"fmt"
	"time"

	"letter-manage-backend/model"

	"gorm.io/gorm"
)

type LetterFilter struct {
	Status        string
	CategoryID    *uint
	Keyword       string
	LetterNo      string
	CitizenName   string
	Phone         string
	IDCard        string
	StartTime     *time.Time
	EndTime       *time.Time
	UnitName      string
	UnitNames     []string
	UnitID        *uint
	UnitIDs       []uint
	HandlerUserID *uint
	HandlerUnitID *uint
	HandlerUnitIDs []uint
	AllUnitID     *uint
	AllUnitIDs    []uint
	Page          int
	PageSize      int
}

// UnitNameToIDs 根据单位名称查找所有匹配的单位 ID
// 短名匹配 level1/level2/level3 任一字段
func UnitNameToIDs(unitName string) []uint {
	shortName := NormalizeUnitName(unitName)
	allUnits, err := GetAllUnits()
	if err != nil {
		return nil
	}
	var ids []uint
	seen := map[uint]bool{}
	for _, u := range allUnits {
		if u.Level1 == shortName || u.Level2 == shortName || u.Level3 == shortName {
			if !seen[u.ID] {
				seen[u.ID] = true
				ids = append(ids, u.ID)
			}
		}
	}
	return ids
}

func buildLetterQuery(filter LetterFilter) *gorm.DB {
	query := DB.Model(&model.Letter{})
	if filter.Status != "" {
		if code, ok := model.StatusNameToCode[filter.Status]; ok {
			query = query.Where("current_status = ?", code)
		}
	}
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
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
	// 单位过滤始终使用 unit_id
	if filter.UnitID != nil {
		query = query.Where("current_unit_id = ?", *filter.UnitID)
	}
	if len(filter.UnitIDs) > 0 {
		query = query.Where("current_unit_id IN ?", filter.UnitIDs)
	}
	// 向后兼容：如果有 UnitName/UnitNames，转为 ID 查询
	if filter.UnitName != "" {
		ids := UnitNameToIDs(filter.UnitName)
		if len(ids) > 0 {
			query = query.Where("current_unit_id IN ?", ids)
		}
	}
	if len(filter.UnitNames) > 0 {
		var allIDs []uint
		seen := map[uint]bool{}
		for _, name := range filter.UnitNames {
			ids := UnitNameToIDs(name)
			for _, id := range ids {
				if !seen[id] {
					seen[id] = true
					allIDs = append(allIDs, id)
				}
			}
		}
		if len(allIDs) > 0 {
			query = query.Where("current_unit_id IN ?", allIDs)
		}
	}
	if filter.HandlerUserID != nil {
		query = query.Where("handler_user_id = ?", *filter.HandlerUserID)
	}
	if filter.HandlerUnitID != nil && filter.AllUnitID != nil {
		query = query.Where("(handler_unit_id = ? OR current_unit_id = ?)", *filter.HandlerUnitID, *filter.AllUnitID)
	} else if filter.HandlerUnitID != nil {
		query = query.Where("handler_unit_id = ?", *filter.HandlerUnitID)
	} else if filter.AllUnitID != nil {
		query = query.Where("current_unit_id = ?", *filter.AllUnitID)
	}
	if len(filter.HandlerUnitIDs) > 0 && len(filter.AllUnitIDs) > 0 {
		query = query.Where("(handler_unit_id IN ? OR current_unit_id IN ?)", filter.HandlerUnitIDs, filter.AllUnitIDs)
	} else if len(filter.HandlerUnitIDs) > 0 {
		query = query.Where("handler_unit_id IN ?", filter.HandlerUnitIDs)
	} else if len(filter.AllUnitIDs) > 0 {
		query = query.Where("current_unit_id IN ?", filter.AllUnitIDs)
	}
	return query
}

func GetLetterList(filter LetterFilter) ([]model.Letter, int64, error) {
	var letters []model.Letter
	var total int64
	query := buildLetterQuery(filter).Preload("Category").Preload("CurrentUnitObj")
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
	err := DB.Where("letter_no = ?", letterNo).Preload("Category").First(&letter).Error
	if err != nil {
		return nil, err
	}
	return &letter, nil
}

func GetLetterByID(id uint) (*model.Letter, error) {
	var letter model.Letter
	err := DB.Where("id = ?", id).Preload("Category").First(&letter).Error
	if err != nil {
		return nil, err
	}
	return &letter, nil
}

func GetLettersByPhone(phone string) ([]model.Letter, error) {
	var letters []model.Letter
	err := DB.Where("phone = ?", phone).Preload("Category").Order("created_at DESC").Find(&letters).Error
	return letters, err
}

func GetLettersByIDCard(idCard string) ([]model.Letter, error) {
	var letters []model.Letter
	err := DB.Where("id_card = ?", idCard).Preload("Category").Order("created_at DESC").Find(&letters).Error
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

func UpdateLetterStatus(letterNo, status string, unitID ...*uint) error {
	code, ok := model.StatusNameToCode[status]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	updates := map[string]interface{}{
		"current_status": code,
	}
	if len(unitID) > 0 && unitID[0] != nil {
		updates["current_unit_id"] = *unitID[0]
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
	return DB.Model(&model.LetterFlow{}).Where("letter_no = ?", letterNo).Update("flow_records", flowRecords).Error
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
		att.CreatedAt = time.Now()
		att.UpdatedAt = time.Now()
		return DB.Create(att).Error
	} else if err != nil {
		return err
	}
	att.ID = existing.ID
	if !existing.CreatedAt.IsZero() {
		att.CreatedAt = existing.CreatedAt
	} else {
		att.CreatedAt = time.Now()
	}
	att.UpdatedAt = time.Now()
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
	Status model.StatusCode `json:"status"`
	Count  int64            `json:"count"`
}

func GetLetterStatusStats(startTime, endTime *time.Time, unitNames []string, handlerUserID ...uint) ([]StatusCount, error) {
	var results []StatusCount
	query := DB.Model(&model.Letter{}).
		Select("current_status as status, count(*) as count")
	if startTime != nil {
		query = query.Where("received_at >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("received_at <= ?", endTime)
	}
	if len(unitNames) > 0 {
		unitIDs := unitNamesToIDs(unitNames)
		if len(unitIDs) > 0 {
			query = query.Where("handler_unit_id IN ?", unitIDs)
		}
	}
	if len(handlerUserID) > 0 && handlerUserID[0] > 0 {
		query = query.Where("handler_user_id = ?", handlerUserID[0])
	}
	err := query.Group("current_status").
		Scan(&results).Error
	return results, err
}

type ChannelCount struct {
	Channel model.ChannelCode `json:"channel"`
	Count   int64             `json:"count"`
}

func GetLetterChannelStats(unitNames []string, handlerUserID ...uint) ([]ChannelCount, error) {
	var results []ChannelCount
	query := DB.Model(&model.Letter{}).
		Select("channel, count(*) as count")
	if len(unitNames) > 0 {
		unitIDs := unitNamesToIDs(unitNames)
		if len(unitIDs) > 0 {
			query = query.Where("handler_unit_id IN ?", unitIDs)
		}
	}
	if len(handlerUserID) > 0 && handlerUserID[0] > 0 {
		query = query.Where("handler_user_id = ?", handlerUserID[0])
	}
	err := query.Group("channel").
		Scan(&results).Error
	return results, err
}

type MonthCount struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}

func GetLetterMonthStats(unitNames []string, handlerUserID ...uint) ([]MonthCount, error) {
	var results []MonthCount
	query := DB.Model(&model.Letter{}).
		Select("DATE_FORMAT(received_at, '%Y-%m') as month, count(*) as count")
	if len(unitNames) > 0 {
		unitIDs := unitNamesToIDs(unitNames)
		if len(unitIDs) > 0 {
			query = query.Where("handler_unit_id IN ?", unitIDs)
		}
	}
	if len(handlerUserID) > 0 && handlerUserID[0] > 0 {
		query = query.Where("handler_user_id = ?", handlerUserID[0])
	}
	err := query.Group("month").
		Order("month ASC").
		Limit(12).
		Scan(&results).Error
	return results, err
}

type CategoryCount struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

func GetLetterCategoryStats(startTime, endTime *time.Time, unitNames []string, handlerUserID ...uint) ([]CategoryCount, error) {
	var results []CategoryCount
	query := DB.Model(&model.Letter{}).
		Select("categories.level1 as category, count(*) as count").
		Joins("LEFT JOIN categories ON categories.id = letters.category_id").
		Where("categories.level1 != ''")
	if startTime != nil {
		query = query.Where("received_at >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("received_at <= ?", endTime)
	}
	if len(unitNames) > 0 {
		unitIDs := unitNamesToIDs(unitNames)
		if len(unitIDs) > 0 {
			query = query.Where("handler_unit_id IN ?", unitIDs)
		}
	}
	if len(handlerUserID) > 0 && handlerUserID[0] > 0 {
		query = query.Where("handler_user_id = ?", handlerUserID[0])
	}
	err := query.Group("categories.level1").
		Order("count DESC").
		Limit(10).
		Scan(&results).Error
	return results, err
}

// ByUnitID variants for direct ID-based queries
func GetLetterStatusStatsByUnitIDs(startTime, endTime *time.Time, unitIDs []uint, handlerUserID ...uint) ([]StatusCount, error) {
	var results []StatusCount
	query := DB.Model(&model.Letter{}).
		Select("current_status as status, count(*) as count")
	if startTime != nil {
		query = query.Where("received_at >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("received_at <= ?", endTime)
	}
	if len(unitIDs) > 0 {
		query = query.Where("(handler_unit_id IN ? OR current_unit_id IN ?)", unitIDs, unitIDs)
	}
	if len(handlerUserID) > 0 && handlerUserID[0] > 0 {
		query = query.Where("handler_user_id = ?", handlerUserID[0])
	}
	err := query.Group("current_status").
		Scan(&results).Error
	return results, err
}

func GetLetterChannelStatsByUnitIDs(unitIDs []uint, handlerUserID ...uint) ([]ChannelCount, error) {
	var results []ChannelCount
	query := DB.Model(&model.Letter{}).
		Select("channel, count(*) as count")
	if len(unitIDs) > 0 {
		query = query.Where("(handler_unit_id IN ? OR current_unit_id IN ?)", unitIDs, unitIDs)
	}
	if len(handlerUserID) > 0 && handlerUserID[0] > 0 {
		query = query.Where("handler_user_id = ?", handlerUserID[0])
	}
	err := query.Group("channel").
		Scan(&results).Error
	return results, err
}

func GetLetterMonthStatsByUnitIDs(unitIDs []uint, handlerUserID ...uint) ([]MonthCount, error) {
	var results []MonthCount
	query := DB.Model(&model.Letter{}).
		Select("DATE_FORMAT(received_at, '%Y-%m') as month, count(*) as count")
	if len(unitIDs) > 0 {
		query = query.Where("(handler_unit_id IN ? OR current_unit_id IN ?)", unitIDs, unitIDs)
	}
	if len(handlerUserID) > 0 && handlerUserID[0] > 0 {
		query = query.Where("handler_user_id = ?", handlerUserID[0])
	}
	err := query.Group("month").
		Order("month ASC").
		Limit(12).
		Scan(&results).Error
	return results, err
}

// TrendPoint 趋势图数据点
type TrendPoint struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

// GetLetterTrend 按指定粒度返回趋势数据
// granularity: "hour", "day", "weekday", "month"
// startTime/endTime 为 nil 时不过滤时间
func GetLetterTrend(unitIDs []uint, handlerUserID uint, granularity string, startTime, endTime *time.Time) ([]TrendPoint, error) {
	var format string
	switch granularity {
	case "hour":
		format = "%H:00"
	case "day":
		format = "%m-%d"
	case "weekday":
		format = "%w"
	case "month":
		format = "%Y-%m"
	default:
		format = "%Y-%m"
	}

	var results []TrendPoint
	query := DB.Model(&model.Letter{}).
		Select(fmt.Sprintf("DATE_FORMAT(received_at, '%s') as label, count(*) as count", format))
	if startTime != nil {
		query = query.Where("received_at >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("received_at <= ?", endTime)
	}
	if len(unitIDs) > 0 {
		query = query.Where("(handler_unit_id IN ? OR current_unit_id IN ?)", unitIDs, unitIDs)
	}
	if handlerUserID > 0 {
		query = query.Where("handler_user_id = ?", handlerUserID)
	}
	err := query.Group("label").Order("label ASC").Scan(&results).Error
	return results, err
}

func GetLetterCategoryStatsByUnitIDs(startTime, endTime *time.Time, unitIDs []uint, handlerUserID ...uint) ([]CategoryCount, error) {
	var results []CategoryCount
	query := DB.Model(&model.Letter{}).
		Select("categories.level1 as category, count(*) as count").
		Joins("LEFT JOIN categories ON categories.id = letters.category_id").
		Where("categories.level1 != ''")
	if startTime != nil {
		query = query.Where("received_at >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("received_at <= ?", endTime)
	}
	if len(unitIDs) > 0 {
		query = query.Where("(handler_unit_id IN ? OR current_unit_id IN ?)", unitIDs, unitIDs)
	}
	if len(handlerUserID) > 0 && handlerUserID[0] > 0 {
		query = query.Where("handler_user_id = ?", handlerUserID[0])
	}
	err := query.Group("categories.level1").
		Order("count DESC").
		Limit(10).
		Scan(&results).Error
	return results, err
}

// unitNamesToIDs 批量将单位名称列表转为单位 ID 列表
func unitNamesToIDs(names []string) []uint {
	seen := map[uint]bool{}
	var ids []uint
	for _, name := range names {
		for _, id := range UnitNameToIDs(name) {
			if !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func GetDispatchList(unitID *uint, permLevel string, page, pageSize int) ([]model.Letter, int64, error) {
	query := DB.Model(&model.Letter{}).Preload("Category")
	switch permLevel {
	case "CITY":
		query = query.Where("current_status = ?", model.StatusCodePreProcess)
	case "DISTRICT":
		// 区县局：显示市局下发至本单位及下属单位的信件，待进一步下发
		var unitIDs []uint
		if unitID != nil {
			unitIDs = GetSubordinateUnitIDs(*unitID)
			if len(unitIDs) == 0 {
				unitIDs = []uint{*unitID}
			}
		}
		if len(unitIDs) > 0 {
			query = query.Where(
				"current_unit_id IN ? AND current_status IN ?",
				unitIDs,
				[]model.StatusCode{model.StatusCodePendingDisDispatch, model.StatusCodeCityDispatched},
			)
		} else {
			query = query.Where("1 = 0")
		}
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

func GetProcessingList(unitID *uint, permLevel string, page, pageSize int, handlerUserID ...uint) ([]model.Letter, int64, error) {
	query := DB.Model(&model.Letter{})

	// 市局可看待审核、处理中的信件（不含已下发走的）
	// 区县局可看本单位及下属单位的待处理信件
	// 基层单位只可看本单位已下发的信件
	if permLevel == "CITY" {
		query = query.Where(
			"current_status IN ?",
			[]model.StatusCode{model.StatusCodeProcessing, model.StatusCodePendingDistrictAudit, model.StatusCodePendingCityAudit},
		)
	} else if permLevel == "DISTRICT" {
		// 区县局：handler_unit_id 过滤本单位及下属单位的信件
		var unitIDs []uint
		if unitID != nil {
			unitIDs = GetSubordinateUnitIDs(*unitID)
			if len(unitIDs) == 0 {
				unitIDs = []uint{*unitID}
			}
		}
		if len(unitIDs) > 0 {
			query = query.Where(
				"handler_unit_id IN ? AND current_status IN ?",
				unitIDs,
				[]model.StatusCode{model.StatusCodeProcessing, model.StatusCodeCityDirectDispatch, model.StatusCodePendingDistrictAudit, model.StatusCodeDispatched},
			)
		} else {
			query = query.Where("1 = 0")
		}
	} else {
		// OFFICER：只能看到本单位已下发、处理中、越级下发的信件
		// 使用 handlerUserID 查询用户的精确 unit_id，避免同名单位污染
		if len(handlerUserID) > 0 && handlerUserID[0] > 0 {
			// 通过用户ID找到精确的 unit_id
			var userPolice model.PoliceUser
			if err := DB.First(&userPolice, handlerUserID[0]).Error; err == nil && userPolice.UnitID != nil {
				query = query.Where(
					"current_unit_id = ? AND current_status IN ?",
					*userPolice.UnitID,
					[]model.StatusCode{model.StatusCodeDispatched, model.StatusCodeProcessing, model.StatusCodeCityDirectDispatch},
				)
			} else {
				// 兜底：unitID 为空，返回空结果
				var unitIDs []uint
				if len(unitIDs) > 0 {
					query = query.Where(
						"current_unit_id IN ? AND current_status IN ?",
						unitIDs,
						[]model.StatusCode{model.StatusCodeDispatched, model.StatusCodeProcessing, model.StatusCodeCityDirectDispatch},
					)
				} else {
					query = query.Where("1 = 0")
				}
			}
		} else {
			// 无 handlerUserID，使用 unitID 精确过滤
			if unitID != nil {
				query = query.Where(
					"current_unit_id = ? AND current_status IN ?",
					*unitID,
					[]model.StatusCode{model.StatusCodeDispatched, model.StatusCodeProcessing, model.StatusCodeCityDirectDispatch},
				)
			} else {
				query = query.Where("1 = 0")
			}
		}
	}
	// 处理人过滤
	if len(handlerUserID) > 0 && handlerUserID[0] > 0 {
		if permLevel == "CITY" {
			// CITY：看到分配给我的 + 未分配的信件
			query = query.Where("(handler_user_id = ? OR handler_user_id IS NULL)", handlerUserID[0])
		} else {
			// DISTRICT / OFFICER：只看到下发到本人的信件
			query = query.Where("handler_user_id = ?", handlerUserID[0])
		}
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
	err := query.Preload("Category").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&letters).Error
	return letters, total, err
}

// GetProcessingListByUnitID 根据单位 ID 获取待处理列表
func GetProcessingListByUnitID(unitID uint, permLevel string, page, pageSize int) ([]model.Letter, int64, error) {
	query := DB.Model(&model.Letter{})

	if permLevel == "CITY" {
		query = query.Where(
			"current_status IN ?",
			[]model.StatusCode{model.StatusCodeProcessing, model.StatusCodePendingDistrictAudit, model.StatusCodePendingCityAudit},
		)
	} else if permLevel == "DISTRICT" {
		// 区县局：可见本单位及下属单位的待处理信件（不含尚未下发的已下发状态）
		subUnitIDs := GetSubordinateUnitIDs(unitID)
		if len(subUnitIDs) > 0 {
			query = query.Where(
				"current_unit_id IN ? AND current_status IN ?",
				subUnitIDs,
				[]model.StatusCode{model.StatusCodeProcessing, model.StatusCodeCityDirectDispatch, model.StatusCodePendingDistrictAudit},
			)
		} else {
			query = query.Where(
				"current_unit_id = ? AND current_status IN ?",
				unitID,
				[]model.StatusCode{model.StatusCodeProcessing, model.StatusCodeCityDirectDispatch, model.StatusCodePendingDistrictAudit},
			)
		}
	} else {
		// OFFICER：可见本单位的已下发、处理中、越级下发信件
		query = query.Where(
			"current_unit_id = ? AND current_status IN ?",
			unitID,
			[]model.StatusCode{model.StatusCodeDispatched, model.StatusCodeProcessing, model.StatusCodeCityDirectDispatch},
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
	err := query.Preload("Category").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&letters).Error
	return letters, total, err
}

func GetAuditList(unitID *uint, permLevel string, page, pageSize int) ([]model.Letter, int64, error) {
	query := DB.Model(&model.Letter{})
	switch permLevel {
	case "CITY":
		// 市局：只查看分局已审核上报的（待市局审核 + 待分县局/支队审核），不包含待核查
		query = query.Where("current_status IN ?", []model.StatusCode{model.StatusCodePendingCityAudit, model.StatusCodePendingDistrictAudit})
	case "DISTRICT":
		// 分县局：查看下发至本单位的待核查信件 + 本单位科室已反馈的信件
		var unitIDs []uint
		if unitID != nil {
			unitIDs = GetSubordinateUnitIDs(*unitID)
			if len(unitIDs) == 0 {
				unitIDs = []uint{*unitID}
			}
		}
		if len(unitIDs) > 0 {
			query = query.Where(
				"(current_status = ? AND current_unit_id IN ?) OR (current_status = ? AND current_unit_id IN ?)",
				model.StatusCodePendingVerification, unitIDs,
				model.StatusCodePendingDistrictAudit, unitIDs,
			)
		} else {
			query = query.Where("1 = 0")
		}
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
	err := query.Preload("Category").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&letters).Error
	return letters, total, err
}

// GetAuditListByUnitID 根据单位 ID 获取审核列表
func GetAuditListByUnitID(unitID uint, permLevel string, page, pageSize int) ([]model.Letter, int64, error) {
	query := DB.Model(&model.Letter{})
	switch permLevel {
	case "CITY":
		// 市局：只查看分局已审核上报的（待市局审核 + 待分县局/支队审核），不包含待核查
		query = query.Where("current_status IN ?", []model.StatusCode{model.StatusCodePendingCityAudit, model.StatusCodePendingDistrictAudit})
	case "DISTRICT":
		// 分县局：查看下发至本单位的待核查信件 + 本单位科室已反馈的信件
		query = query.Where(
			"(current_status = ? AND current_unit_id = ?) OR (current_status = ? AND current_unit_id = ?)",
			model.StatusCodePendingVerification, unitID,
			model.StatusCodePendingDistrictAudit, unitID,
		)
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
	err := query.Preload("Category").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&letters).Error
	return letters, total, err
}

// GetLettersForExport 导出用：不限条数，预加载 Category，支持筛选
func GetLettersForExport(filter LetterFilter) ([]model.Letter, error) {
	var letters []model.Letter
	query := DB.Model(&model.Letter{}).Preload("Category")
	if filter.StartTime != nil {
		query = query.Where("received_at >= ?", filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("received_at <= ?", filter.EndTime)
	}
	if filter.Status != "" {
		if code, ok := model.StatusNameToCode[filter.Status]; ok {
			query = query.Where("current_status = ?", code)
		}
	}
	if filter.CategoryID != nil {
		query = query.Where("category_id = ?", *filter.CategoryID)
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
	// 单位过滤
	if len(filter.AllUnitIDs) > 0 {
		query = query.Where("(handler_unit_id IN ? OR current_unit_id IN ?)", filter.AllUnitIDs, filter.AllUnitIDs)
	} else if filter.AllUnitID != nil {
		query = query.Where("(handler_unit_id = ? OR current_unit_id = ?)", *filter.AllUnitID, *filter.AllUnitID)
	}
	if filter.HandlerUserID != nil {
		query = query.Where("handler_user_id = ?", *filter.HandlerUserID)
	}
	err := query.Order("received_at DESC").Find(&letters).Error
	return letters, err
}
