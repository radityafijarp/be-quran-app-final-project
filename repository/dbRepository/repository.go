package dbRepository

import (
	"a21hc3NpZ25tZW50/model"
	"errors"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) AddUser(user model.User) (string, error) {
	err := r.db.Create(&user).Error
	if err != nil {
		return "", err
	}
	return user.Username, nil
}

func (r *Repository) GetUserByUsername(username string) (model.User, error) {
	var user model.User
	result := r.db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return model.User{}, nil // Return empty user if not found
		}
		return model.User{}, result.Error
	}
	return user, nil
}

// Add Memorize record
func (r *Repository) AddMemorize(memorize model.Memorize) (uint, error) {
	err := r.db.Create(&memorize).Error
	if err != nil {
		return 0, err
	}
	return memorize.ID, nil
}

// Get Memorize record by ID
func (r *Repository) GetMemorizeByID(memorizeID uint) (model.Memorize, error) {
	var memorize model.Memorize
	err := r.db.First(&memorize, memorizeID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Memorize{}, nil
		}
		return model.Memorize{}, err
	}
	return memorize, nil
}

// Delete Memorize record by ID
func (r *Repository) DeleteMemorize(memorizeID uint) error {
	result := r.db.Delete(&model.Memorize{}, memorizeID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("memorize record not found")
	}
	return nil
}

// Get all Memorize records for a user
func (r *Repository) GetAllMemorizesByUser(username string) ([]model.Memorize, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No user found, return empty slice
		}
		return nil, err
	}

	var memorizes []model.Memorize
	err = r.db.Where("user_id = ?", user.ID).Find(&memorizes).Error
	if err != nil {
		return nil, err
	}
	return memorizes, nil
}

func (r *Repository) UpdateMemorize(memorize model.Memorize) error {
	return r.db.Save(&memorize).Error
}
