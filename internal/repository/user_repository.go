package repository

import (
	"vestra-ecommerce-backend/internal/domain"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	Update(user *domain.User) error
    
    FindByID(id uuid.UUID) (*domain.User, error)
    FindAll() ([]domain.User, error)
    SoftDelete(id uuid.UUID) error
}
