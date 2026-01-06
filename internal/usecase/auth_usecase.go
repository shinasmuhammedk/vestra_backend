package usecase

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"vestra-ecommerce-backend/internal/domain"
	"vestra-ecommerce-backend/internal/repository"
	"vestra-ecommerce-backend/internal/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase handles authentication-related business logic
type AuthUseCase struct {
	userRepo repository.UserRepository
	otpRepo  repository.OTPRepository
}

// Constructor for AuthUseCase
func NewAuthUseCase(userRepo repository.UserRepository, otpRepo repository.OTPRepository) *AuthUseCase {
	rand.Seed(time.Now().UnixNano()) // seed random once
	return &AuthUseCase{
		userRepo: userRepo,
		otpRepo:  otpRepo,
	}
}

// RegisterUser registers a new user and sends OTP via email
func (a *AuthUseCase) RegisterUser(name, email, password, role string) error {
	// 1️⃣ Check if user already exists
	existingUser, _ := a.userRepo.FindByEmail(email)
	if existingUser != nil {
		return errors.New("email already registered")
	}

	// 2️⃣ Hash password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	// 3️⃣ Create user (unverified)
	user := &domain.User{
		ID:         uuid.New(),
		Name:       name,
		Email:      email,
		Password:   hashedPassword,
		Role:       role,
		IsVerified: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = a.userRepo.Create(user)
	if err != nil {
		return err
	}

	// 4️⃣ Generate OTP
	otpCode := generateOTP()
	otp := &domain.OTP{
		ID:        uuid.New(),
		UserID:    user.ID,
		Code:      otpCode,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
		CreatedAt: time.Now(),
	}

	// Save OTP in DB
	err = a.otpRepo.Save(otp)
	if err != nil {
		return err
	}

	// 5️⃣ Send OTP via email (real)
	err = utils.SendOTPEmail(user.Email, otpCode)
	if err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	return nil
}

// VerifyOTP validates the OTP and marks user as verified
func (a *AuthUseCase) VerifyOTP(email, code string) error {
	// 1️⃣ Find user
	user, err := a.userRepo.FindByEmail(email)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// 2️⃣ Prevent re-verification
	if user.IsVerified {
		return errors.New("user already verified")
	}

	// 3️⃣ Find valid OTP
	otp, err := a.otpRepo.FindValidOTP(user.ID, code)
	if err != nil || otp == nil {
		return errors.New("invalid or expired OTP")
	}

	// 4️⃣ Mark OTP as used
	err = a.otpRepo.MarkUsed(otp.ID)
	if err != nil {
		return err
	}

	// 5️⃣ Mark user as verified
	user.IsVerified = true
	user.UpdatedAt = time.Now()
	err = a.userRepo.Update(user)
	if err != nil {
		return err
	}

	return nil
}

///////////////////////
// Helper Functions  //
///////////////////////

// hashPassword hashes a plain-text password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// generateOTP generates a random 6-digit OTP
func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
