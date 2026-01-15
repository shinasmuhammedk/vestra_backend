// 
package model

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`

	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email" gorm:"uniqueIndex"`

	Password string `json:"-" validate:"required,min=8"`

	Role string `json:"role" gorm:"default:user"`

	OTP       string    `json:"-" gorm:"size:255"`
	OTPExpiry time.Time `json:"-"`

	IsVerified bool `json:"is_verified" gorm:"default:false"`
	IsBlocked  bool `json:"is_blocked" gorm:"default:false"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Automatically generate UUID before creating user
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Hash password before saving (only if not already hashed)
func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	if u.Password != "" && !isHashed(u.Password) {
		hashed, err := bcrypt.GenerateFromPassword(
			[]byte(u.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return err
		}
		u.Password = string(hashed)
	}
	return nil
}

// Check if password is already bcrypt-hashed
func isHashed(password string) bool {
	if len(password) < 60 {
		return false
	}

	prefix := password[:4]
	return prefix == "$2a$" || prefix == "$2b$" || prefix == "$2y$"
}
