package repository

import "vestra-ecommerce-backend/internal/domain"

type UserRepository interface {
	Create(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	Update(user *domain.User) error
}
