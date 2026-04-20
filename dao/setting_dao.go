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
