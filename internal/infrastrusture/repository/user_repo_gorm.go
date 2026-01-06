package repository

import (
	"vestra-ecommerce-backend/internal/domain"
	"vestra-ecommerce-backend/internal/repository"

	"gorm.io/gorm"
)

type UserRepoGorm struct {
	db *gorm.DB
}

func NewUserRepoGorm(db *gorm.DB) repository.UserRepository {
	return &UserRepoGorm{db: db}
}

func (r *UserRepoGorm) Create(user *domain.User) error {
	model := UserModel{
		ID:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		Password:   user.Password,
		Role:       user.Role,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
	return r.db.Create(&model).Error
}

func (r *UserRepoGorm) FindByEmail(email string) (*domain.User, error) {
	var model UserModel
	if err := r.db.Where("email = ?", email).First(&model).Error; err != nil {
		return nil, err
	}

	return &domain.User{
		ID:         model.ID,
		Name:       model.Name,
		Email:      model.Email,
		Password:   model.Password,
		Role:       model.Role,
		IsVerified: model.IsVerified,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}, nil
}

func (r *UserRepoGorm) Update(user *domain.User) error {
	return r.db.Model(&UserModel{}).Where("id = ?", user.ID).Updates(UserModel{
		Name:       user.Name,
		Email:      user.Email,
		Password:   user.Password,
		Role:       user.Role,
		IsVerified: user.IsVerified,
		UpdatedAt:  user.UpdatedAt,
	}).Error
}
