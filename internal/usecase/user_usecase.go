package usecase

import (
	"errors"

	"vestra-ecommerce-backend/internal/domain"
	"vestra-ecommerce-backend/internal/repository"

	"github.com/google/uuid"
)

type UserUseCase struct {
	userRepo repository.UserRepository
}

func NewUserUseCase(userRepo repository.UserRepository) *UserUseCase {
	return &UserUseCase{userRepo: userRepo}
}

func (u *UserUseCase) GetProfile(userID string) (*domain.User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	user, err := u.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.Password = "" // 🔐 never expose password
	return user, nil
}
