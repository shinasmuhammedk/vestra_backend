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

// func (a *AuthUseCase) LoginUser(email, password string) (string, error) {
// 	// 1️⃣ Find user
// 	user, err := a.userRepo.FindByEmail(email)
// 	if err != nil || user == nil {
// 		return "", errors.New("invalid email or password")
// 	}

// 	// 2️⃣ Check if verified
// 	if !user.IsVerified {
// 		return "", errors.New("please verify your email before login")
// 	}

// 	// 3️⃣ Compare password
// 	err = bcrypt.CompareHashAndPassword(
// 		[]byte(user.Password),
// 		[]byte(password),
// 	)
// 	if err != nil {
// 		return "", errors.New("invalid email or password")
// 	}

// 	// 4️⃣ Generate JWT token
// 	token, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role)
// 	if err != nil {
// 		return "", err
// 	}

// 	return token, nil
// }

func (a *AuthUseCase) LoginUser(email, password string) (string, string, error) {
	// 1️⃣ Find user by email
	user, err := a.userRepo.FindByEmail(email)
	if err != nil || user == nil {
		return "", "", errors.New("invalid email or password")
	}
    
    //Check blocked or not
	if user.IsBlocked == true{
		return "", "", errors.New("account is blocked by admin")
	}

	// 2️⃣ Check email verification
	if !user.IsVerified {
		return "", "", errors.New("please verify your email before login")
	}

	// 3️⃣ Compare password
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)
	if err != nil {
		return "", "", errors.New("invalid email or password")
	}

	// 4️⃣ Generate access token
	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Role,
	)
	if err != nil {
		return "", "", err
	}

	// 5️⃣ Generate refresh token (FIXED)
	refreshToken, _, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// =========Forgot password======
// ForgotPassword sends OTP for password reset
func (a *AuthUseCase) ForgotPassword(email string) error {
	// 1️⃣ Check user exists
	user, err := a.userRepo.FindByEmail(email)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// 2️⃣ Generate OTP
	otpCode := generateOTP()
	otp := &domain.OTP{
		ID:        uuid.New(),
		UserID:    user.ID,
		Code:      otpCode,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
		CreatedAt: time.Now(),
	}

	// 3️⃣ Save OTP
	if err := a.otpRepo.Save(otp); err != nil {
		return err
	}

	// 4️⃣ Send OTP email
	if err := utils.SendOTPEmail(user.Email, otpCode); err != nil {
		return err
	}

	return nil
}

// =================RESET PASSWORD===============
// ResetPassword resets user password using OTP
func (a *AuthUseCase) ResetPassword(email, otpCode, newPassword string) error {
	// 1️⃣ Find user
	user, err := a.userRepo.FindByEmail(email)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// 2️⃣ Validate OTP
	otp, err := a.otpRepo.FindValidOTP(user.ID, otpCode)
	if err != nil || otp == nil {
		return errors.New("invalid or expired OTP")
	}

	// 3️⃣ Hash new password
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	// 4️⃣ Update password
	user.Password = hashedPassword
	user.UpdatedAt = time.Now()

	if err := a.userRepo.Update(user); err != nil {
		return err
	}

	// 5️⃣ Mark OTP used
	if err := a.otpRepo.MarkUsed(otp.ID); err != nil {
		return err
	}

	return nil
}

func (a *AuthUseCase) RefreshAccessToken(refreshToken string) (string, error) {
	// 1️⃣ Validate refresh token
	claims, err := utils.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", errors.New("invalid or expired refresh token")
	}

	// 2️⃣ Parse user ID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", errors.New("invalid token user")
	}

	// 3️⃣ Fetch user
	user, err := a.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return "", errors.New("user njjjot found")
	}

	// 4️⃣ Generate new access token
	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
