package dao

import (
	"letter-manage-backend/model"
)

// Category DAO

func GetAllCategories() ([]model.Category, error) {
	var cats []model.Category
	err := DB.Find(&cats).Error
	return cats, err
}

func GetCategoryByID(id uint) (*model.Category, error) {
	var cat model.Category
	err := DB.First(&cat, id).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func CreateCategory(cat *model.Category) error {
	return DB.Create(cat).Error
}

func UpdateCategory(cat *model.Category) error {
	return DB.Save(cat).Error
}

func DeleteCategory(id uint) error {
	return DB.Delete(&model.Category{}, id).Error
}

// SpecialFocus DAO

func GetAllSpecialFocuses() ([]model.SpecialFocus, error) {
	var sfs []model.SpecialFocus
	err := DB.Find(&sfs).Error
	return sfs, err
}

func GetSpecialFocusByID(id uint) (*model.SpecialFocus, error) {
	var sf model.SpecialFocus
	err := DB.First(&sf, id).Error
	if err != nil {
		return nil, err
	}
	return &sf, nil
}

func CreateSpecialFocus(sf *model.SpecialFocus) error {
	return DB.Create(sf).Error
}

func UpdateSpecialFocus(sf *model.SpecialFocus) error {
	return DB.Save(sf).Error
}

func DeleteSpecialFocus(id uint) error {
	return DB.Delete(&model.SpecialFocus{}, id).Error
}

// Prompt DAO

func GetPromptByType(promptType string) (*model.Prompt, error) {
	var p model.Prompt
	err := DB.Where("prompt_type = ?", promptType).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func GetAllPrompts() ([]model.Prompt, error) {
	var prompts []model.Prompt
	err := DB.Find(&prompts).Error
	return prompts, err
}

// ──── LetterSpecialFocus DAO ────

// AddLetterSpecialFocus 添加信件-专项关注绑定
func AddLetterSpecialFocus(letterNo string, focusID uint) error {
	lsf := &model.LetterSpecialFocus{LetterNo: letterNo, FocusID: focusID}
	return DB.Create(lsf).Error
}

// RemoveLetterSpecialFocusesByLetterNo 删除信件的所有专项关注绑定
func RemoveLetterSpecialFocusesByLetterNo(letterNo string) error {
	return DB.Where("letter_no = ?", letterNo).Delete(&model.LetterSpecialFocus{}).Error
}

// GetFocusIDsByLetterNo 获取信件绑定的专项关注 ID 列表
func GetFocusIDsByLetterNo(letterNo string) ([]uint, error) {
	var lsfs []model.LetterSpecialFocus
	if err := DB.Where("letter_no = ?", letterNo).Find(&lsfs).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, len(lsfs))
	for i, lsf := range lsfs {
		ids[i] = lsf.FocusID
	}
	return ids, nil
}

// CountLettersByFocusID 统计绑定了指定专项关注的去重信件数量
func CountLettersByFocusID(focusID uint) int64 {
	var count int64
	DB.Model(&model.LetterSpecialFocus{}).
		Where("focus_id = ?", focusID).
		Distinct("letter_no").
		Count(&count)
	return count
}

// GetLetterNosByFocusID 获取绑定了指定专项关注的信件编号列表
func GetLetterNosByFocusID(focusID uint) ([]string, error) {
	var lsfs []model.LetterSpecialFocus
	if err := DB.Where("focus_id = ?", focusID).Distinct("letter_no").Find(&lsfs).Error; err != nil {
		return nil, err
	}
	nos := make([]string, len(lsfs))
	for i, lsf := range lsfs {
		nos[i] = lsf.LetterNo
	}
	return nos, nil
}

// GetFocusIDsByLetterNos 批量获取多条信件的 focus_id (返回 map[letter_no]focus_id)
func GetFocusIDsByLetterNos(letterNos []string) (map[string]uint, error) {
	var lsfs []model.LetterSpecialFocus
	if err := DB.Where("letter_no IN ?", letterNos).Find(&lsfs).Error; err != nil {
		return nil, err
	}
	result := make(map[string]uint, len(lsfs))
	// 每条 letter_no 只保留最近一条
	for _, lsf := range lsfs {
		result[lsf.LetterNo] = lsf.FocusID
	}
	return result, nil
}
