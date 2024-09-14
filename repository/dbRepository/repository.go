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

func (r *Repository) AddPhoto(photo model.Photo) (uint, error) {
	err := r.db.Create(&photo).Error
	if err != nil {
		return 0, err
	}
	return photo.ID, nil
}

func (r *Repository) GetPhotoByID(photoID uint) (model.Photo, error) {
	var photo model.Photo
	err := r.db.First(&photo, photoID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Photo{}, nil
		}
		return model.Photo{}, err
	}
	return photo, nil
}

func (r *Repository) DeletePhoto(photoID uint) error {
	result := r.db.Delete(&model.Photo{}, photoID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("photo not found")
	}
	return nil
}

func (r *Repository) GetAllPhotosByUser(username string) ([]model.Photo, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No user found, return empty slice
		}
		return nil, err
	}

	var photos []model.Photo
	err = r.db.Where("user_id = ?", user.ID).Find(&photos).Error
	if err != nil {
		return nil, err
	}
	return photos, nil
}
