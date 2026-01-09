package usecase

import (
	"errors"

	"vestra-ecommerce-backend/internal/domain"
	"vestra-ecommerce-backend/internal/repository"

	"github.com/google/uuid"
)

type AdminUsecase struct {
	userRepo repository.UserRepository
}

func NewAdminUsecase(userRepo repository.UserRepository) *AdminUsecase {
	return &AdminUsecase{
		userRepo: userRepo,
	}
}

// SetBlockStatus blocks or unblocks a user
func (uc *AdminUsecase) SetBlockStatus(userID uuid.UUID, block bool) error {

	// 1️⃣ Find user
	user, err := uc.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// 2️⃣ Update block status
	user.IsBlocked = block

	// 3️⃣ Save
	return uc.userRepo.Update(user)
}


func (uc *AdminUsecase) GetAllUsers() ([]domain.User, error) {
	return uc.userRepo.FindAll()
}



func (uc *AdminUsecase) SoftDeleteUser(userID uuid.UUID) error {
    user, err := uc.userRepo.FindByID(userID)
    if err != nil {
        return errors.New("user not found")
    }

    return uc.userRepo.SoftDelete(user.ID)
}
