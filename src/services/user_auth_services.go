package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"vestra-ecommerce/src/model"
	"vestra-ecommerce/src/repo"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/email"
	"vestra-ecommerce/utils/utils/apperror"
)

// Seed OTP generator once
func init() {
	rand.Seed(time.Now().UnixNano())
}

type UserAuthService struct {
	userRepo  repo.IPgSQLRepository
	otpExpiry time.Duration
}

// ✅ FIXED: use injected repository (NO globals)
func NewUserAuthService(userRepo repo.IPgSQLRepository, otpExpiryMinutes int) *UserAuthService {
	return &UserAuthService{
		userRepo:  userRepo,
		otpExpiry: time.Duration(otpExpiryMinutes) * time.Minute,
	}
}

// Signup creates user + OTP + sends email
func (s *UserAuthService) Signup(name, userEmail, password string) error {
	var existing model.User

	// Check if email exists
	if err := s.userRepo.FindOneWhere(&existing, "email = ?", userEmail); err == nil {
		return apperror.New(
			constant.BADREQUEST,
            "",
			"Email already exists",
		)
	}

	// Generate OTP
	otp := generateOTP()
	otpHash, _ := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)

	// Create user (password will be hashed by GORM hook)
	user := model.User{
		ID:         uuid.New(),
		Name:       name,
		Email:      userEmail,
		Password:   password, // plain text → hashed in model hook
		Role:       "user",
		OTP:        string(otpHash),
		OTPExpiry:  time.Now().Add(5 * time.Minute),
		IsVerified: false,
	}

	if err := s.userRepo.Insert(&user); err != nil {
		return err
	}

	// Send OTP email
	if err := email.SendOTP(userEmail, otp); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to send OTP email",
		)
	}

	return nil
}

// VerifyOTP validates OTP and activates account
func (s *UserAuthService) VerifyOTP(userEmail, otp string) error {
	var user model.User

	if err := s.userRepo.FindOneWhere(&user, "email = ?", userEmail); err != nil {
		return apperror.New(
			constant.NOTFOUND,
			"",
			"User not found",
		)
	}

	if user.IsVerified {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"User already verified",
		)
	}

	if time.Now().After(user.OTPExpiry) {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"OTP expired",
		)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.OTP), []byte(otp)); err != nil {
		return apperror.New(
			constant.UNAUTHORIZED,
			"",
			"Invalid OTP",
		)
	}

	updates := map[string]interface{}{
		"is_verified": true,
		"otp":         "",
		"otp_expiry":  time.Time{},
	}

	return s.userRepo.UpdateByFields(&model.User{}, user.ID, updates)
}

func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *UserAuthService) Login(email, password string) (*model.User, error) {
	var user model.User

	// 1. Find user by email
	if err := s.userRepo.FindOneWhere(&user, "email = ?", email); err != nil {
		return nil, apperror.New(
			constant.UNAUTHORIZED,
			"",
			"Invalid email or password",
		)
	}

	// 2. Check if verified
	if !user.IsVerified {
		return nil, apperror.New(
			constant.UNAUTHORIZED,
			"",
			"User not verified",
		)
	}

	// 3. Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, apperror.New(
			constant.UNAUTHORIZED,
			"",
			"Invalid email or password",
		)
	}

	// 4. Return user (tokens generated in controller)
	return &user, nil
}

func (s *UserAuthService) ForgotPassword(userEmail string) error {
	var user model.User

	if err := s.userRepo.FindOneWhere(&user, "email = ?", userEmail); err != nil {
		return nil // always return nil for security
	}

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	otpHash, _ := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)

	updates := map[string]interface{}{
		"otp":        string(otpHash),
		"otp_expiry": time.Now().Add(15 * time.Minute),
	}
	if err := s.userRepo.UpdateByFields(&model.User{}, user.ID, updates); err != nil {
		return err
	}

	// Call package function
	if err := email.SendOTP(user.Email, otp); err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to send OTP email",
		)
	}

	return nil
}

func (s *UserAuthService) ResetPassword(
	email string,
	otp string,
	newPassword string,
) error {
	var user model.User

	// 1️⃣ Find user
	if err := s.userRepo.FindOneWhere(&user, "email = ?", email); err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid email or OTP",
		)
	}

	// 2️⃣ Check OTP expiry
	if time.Now().After(user.OTPExpiry) {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"OTP expired",
		)
	}

	// 3️⃣ Compare OTP
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.OTP),
		[]byte(otp),
	); err != nil {
		return apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid OTP",
		)
	}

	// 4️⃣ Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(newPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to hash password",
		)
	}

	// 5️⃣ Update password + clear OTP
	updates := map[string]interface{}{
		"password":   string(hashedPassword),
		"otp":        "",
		"otp_expiry": time.Time{},
	}

	if err := s.userRepo.UpdateByFields(&model.User{}, user.ID, updates); err != nil {
		return err
	}

	return nil
}

func (s *UserAuthService) GetProfile(userID string) (*model.User, error) {
	var user model.User

	if err := s.userRepo.FindById(&user, userID); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"User not found",
		)
	}

	return &user, nil
}

func (s *UserAuthService) UpdateProfile(userID string, name string) (*model.User, error) {
	var user model.User

	if err := s.userRepo.FindById(&user, userID); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"User not found",
		)
	}

	updates := map[string]interface{}{
		"name": name,
	}

	if err := s.userRepo.UpdateByFields(&model.User{}, user.ID, updates); err != nil {
		return nil, err
	}

	// Reload updated user
	if err := s.userRepo.FindById(&user, userID); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserAuthService) ToggleUserBlock(userID string) (*model.User, error) {
	var user model.User

	// 1️⃣ Find user by ID
	if err := s.userRepo.FindById(&user, userID); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"User not found",
		)
	}

	// 2️⃣ Toggle the is_blocked field
	newStatus := !user.IsBlocked
	updates := map[string]interface{}{
		"is_blocked": newStatus,
	}

	if err := s.userRepo.UpdateByFields(&model.User{}, user.ID, updates); err != nil {
		return nil, err
	}

	// 3️⃣ Reload updated user
	if err := s.userRepo.FindById(&user, userID); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByID fetches a user by ID
func (s *UserAuthService) GetByID(userID string) (*model.User, error) {
	var user model.User

	if err := s.userRepo.FindById(&user, userID); err != nil {
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"User not found",
		)
	}

	return &user, nil
}
