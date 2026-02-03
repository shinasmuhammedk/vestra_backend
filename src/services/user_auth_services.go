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
	"vestra-ecommerce/utils/logging" // Import logging package
	"vestra-ecommerce/utils/utils/apperror"
)

// init initializes the random seed for OTP generation
func init() {
	rand.Seed(time.Now().UnixNano())
	// logging.Debug.Println("Random seed initialized for OTP generation")
}

// UserAuthService handles all authentication-related business logic
type UserAuthService struct {
	userRepo  repo.IPgSQLRepository // Repository interface for database operations
	otpExpiry time.Duration         // OTP expiration duration
}

// NewUserAuthService creates a new instance of UserAuthService with dependency injection
// Parameters:
//   - userRepo: Repository implementation for database operations
//   - otpExpiryMinutes: OTP expiration time in minutes
//
// Returns:
//   - *UserAuthService: Initialized user authentication service
func NewUserAuthService(userRepo repo.IPgSQLRepository, otpExpiryMinutes int) *UserAuthService {
	logging.Debug.Printf("Initializing UserAuthService with OTP expiry: %d minutes", otpExpiryMinutes)
	return &UserAuthService{
		userRepo:  userRepo,
		otpExpiry: time.Duration(otpExpiryMinutes) * time.Minute,
	}
}

// Signup handles user registration by creating a new user, generating OTP, and sending verification email
// Parameters:
//   - name: User's full name
//   - userEmail: User's email address
//   - password: User's plain text password (will be hashed)
//
// Returns:
//   - error: Returns an error if registration fails, nil on success
func (s *UserAuthService) Signup(name, userEmail, password string) error {
	logging.Debug.Printf("Starting signup process for email: %s", userEmail)

	var existing model.User

	// Check if email already exists in the database
	if err := s.userRepo.FindOneWhere(&existing, "email = ?", userEmail); err == nil {
		logging.Error.Printf("Signup failed - Email already exists: %s", userEmail)
		return apperror.New(
			constant.BADREQUEST,
			"",
			"Email already exists",
		)
	}
	logging.Debug.Println("Email check passed - email not registered")

	// Generate 6-digit OTP
	otp := generateOTP()
	logging.Debug.Printf("Generated OTP for user: %s", userEmail)
	
	// Hash OTP for secure storage
	otpHash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		logging.Error.Printf("Failed to hash OTP for user %s: %v", userEmail, err)
		return err
	}

	// Create new user object
	user := model.User{
		ID:         uuid.New(),
		Name:       name,
		Email:      userEmail,
		Password:   password, // Plain text - will be hashed by GORM BeforeSave hook
		Role:       "user",
		OTP:        string(otpHash),
		OTPExpiry:  time.Now().Add(s.otpExpiry),
		IsVerified: false,
	}
	logging.Debug.Printf("Created user object with ID: %s", user.ID.String())

	// Insert user into database
	if err := s.userRepo.Insert(&user); err != nil {
		logging.Error.Printf("Failed to insert user %s into database: %v", userEmail, err)
		return err
	}
	logging.Debug.Printf("User inserted successfully: %s", userEmail)

	// Send OTP verification email
	if err := email.SendOTP(userEmail, otp); err != nil {
		logging.Error.Printf("Failed to send OTP email to %s: %v", userEmail, err)
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to send OTP email",
		)
	}
	logging.Debug.Printf("OTP email sent successfully to: %s", userEmail)
	logging.Debug.Printf("Signup process completed for: %s", userEmail)

	return nil
}

// VerifyOTP validates the OTP and activates the user account
// Parameters:
//   - userEmail: User's email address
//   - otp: One-Time Password entered by user
//
// Returns:
//   - error: Returns an error if verification fails, nil on success
func (s *UserAuthService) VerifyOTP(userEmail, otp string) error {
	logging.Debug.Printf("Starting OTP verification for email: %s", userEmail)

	var user model.User

	// Find user by email
	if err := s.userRepo.FindOneWhere(&user, "email = ?", userEmail); err != nil {
		logging.Error.Printf("User not found during OTP verification: %s", userEmail)
		return apperror.New(
			constant.NOTFOUND,
			"",
			"User not found",
		)
	}
	logging.Debug.Printf("User found for OTP verification: %s", userEmail)

	// Check if user is already verified
	if user.IsVerified {
		logging.Debug.Printf("User already verified: %s", userEmail)
		return apperror.New(
			constant.BADREQUEST,
			"",
			"User already verified",
		)
	}

	// Check OTP expiry
	if time.Now().After(user.OTPExpiry) {
		logging.Error.Printf("OTP expired for user: %s", userEmail)
		return apperror.New(
			constant.BADREQUEST,
			"",
			"OTP expired",
		)
	}
	logging.Debug.Println("OTP expiry check passed")

	// Compare hashed OTP with provided OTP
	if err := bcrypt.CompareHashAndPassword([]byte(user.OTP), []byte(otp)); err != nil {
		logging.Error.Printf("Invalid OTP provided for user: %s", userEmail)
		return apperror.New(
			constant.UNAUTHORIZED,
			"",
			"Invalid OTP",
		)
	}
	logging.Debug.Printf("OTP validation successful for: %s", userEmail)

	// Prepare updates to mark user as verified and clear OTP
	updates := map[string]interface{}{
		"is_verified": true,
		"otp":         "",
		"otp_expiry":  time.Time{},
	}

	// Update user in database
	if err := s.userRepo.UpdateByFields(&model.User{}, user.ID, updates); err != nil {
		logging.Error.Printf("Failed to update user verification status: %s - %v", userEmail, err)
		return err
	}
	
	logging.Debug.Printf("User verified successfully: %s", userEmail)
	return nil
}

// generateOTP generates a random 6-digit OTP
// Returns:
//   - string: 6-digit OTP as string
func generateOTP() string {
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	logging.Debug.Printf("Generated OTP: %s", otp)
	return otp
}

// Login authenticates a user and returns user details if successful
// Parameters:
//   - email: User's email address
//   - password: User's password
//
// Returns:
//   - *model.User: User object if authentication succeeds
//   - error: Error if authentication fails
func (s *UserAuthService) Login(email, password string) (*model.User, error) {
	logging.Debug.Printf("Login attempt for email: %s", email)

	var user model.User

	// Find user by email
	if err := s.userRepo.FindOneWhere(&user, "email = ?", email); err != nil {
		logging.Error.Printf("Login failed - User not found: %s", email)
		return nil, apperror.New(
			constant.UNAUTHORIZED,
			"",
			"Invalid email or password",
		)
	}
	logging.Debug.Printf("User found for login: %s", email)

	// Check if user is verified
	if !user.IsVerified {
		logging.Error.Printf("Login failed - User not verified: %s", email)
		return nil, apperror.New(
			constant.UNAUTHORIZED,
			"",
			"User not verified",
		)
	}
	logging.Debug.Println("User verification check passed")

	// Check if user is blocked
	if user.IsBlocked {
		logging.Error.Printf("Login failed - User account blocked: %s", email)
		return nil, apperror.New(
			constant.FORBIDDEN,
			"",
			"Your account has been blocked. Contact support.",
		)
	}
	logging.Debug.Println("User account status check passed")

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logging.Error.Printf("Login failed - Invalid password for: %s", email)
		return nil, apperror.New(
			constant.UNAUTHORIZED,
			"",
			"Invalid email or password",
		)
	}
	logging.Debug.Printf("Password verification successful for: %s", email)
	logging.Debug.Printf("Login successful for user: %s", email)

	return &user, nil
}

// ForgotPassword initiates the password reset process by generating and sending OTP
// Parameters:
//   - userEmail: User's email address
//
// Returns:
//   - error: Returns an error if process fails, always returns nil for security
func (s *UserAuthService) ForgotPassword(userEmail string) error {
	logging.Debug.Printf("Forgot password process initiated for: %s", userEmail)

	var user model.User

	// Find user by email
	if err := s.userRepo.FindOneWhere(&user, "email = ?", userEmail); err != nil {
		// Return nil for security (don't reveal if user exists)
		logging.Debug.Printf("Forgot password - No user found with email: %s (security measure)", userEmail)
		return nil
	}
	logging.Debug.Printf("User found for password reset: %s", userEmail)

	// Generate new OTP
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	logging.Debug.Printf("Generated reset OTP for user: %s", userEmail)
	
	// Hash OTP
	otpHash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		logging.Error.Printf("Failed to hash reset OTP for user %s: %v", userEmail, err)
		return err
	}

	// Prepare updates for OTP and expiry
	updates := map[string]interface{}{
		"otp":        string(otpHash),
		"otp_expiry": time.Now().Add(15 * time.Minute),
	}
	
	// Update user record with new OTP
	if err := s.userRepo.UpdateByFields(&model.User{}, user.ID, updates); err != nil {
		logging.Error.Printf("Failed to update user with reset OTP: %s - %v", userEmail, err)
		return err
	}
	logging.Debug.Printf("Reset OTP updated in database for: %s", userEmail)

	// Send OTP email
	if err := email.SendOTP(user.Email, otp); err != nil {
		logging.Error.Printf("Failed to send reset OTP email to %s: %v", userEmail, err)
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to send OTP email",
		)
	}
	logging.Debug.Printf("Reset OTP email sent successfully to: %s", userEmail)
	logging.Debug.Printf("Forgot password process completed for: %s", userEmail)

	return nil
}

// ResetPassword validates OTP and updates user's password
// Parameters:
//   - email: User's email address
//   - otp: One-Time Password for verification
//   - newPassword: New password to set
//
// Returns:
//   - error: Returns an error if reset fails, nil on success
func (s *UserAuthService) ResetPassword(email string, otp string, newPassword string) error {
	logging.Debug.Printf("Reset password process started for: %s", email)

	var user model.User

	// Find user by email
	if err := s.userRepo.FindOneWhere(&user, "email = ?", email); err != nil {
		logging.Error.Printf("Reset password failed - User not found: %s", email)
		return apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid email or OTP",
		)
	}
	logging.Debug.Printf("User found for password reset: %s", email)

	// Check OTP expiry
	if time.Now().After(user.OTPExpiry) {
		logging.Error.Printf("Reset password failed - OTP expired for: %s", email)
		return apperror.New(
			constant.BADREQUEST,
			"",
			"OTP expired",
		)
	}
	logging.Debug.Println("OTP expiry check passed")

	// Verify OTP
	if err := bcrypt.CompareHashAndPassword([]byte(user.OTP), []byte(otp)); err != nil {
		logging.Error.Printf("Reset password failed - Invalid OTP for: %s", email)
		return apperror.New(
			constant.BADREQUEST,
			"",
			"Invalid OTP",
		)
	}
	logging.Debug.Printf("OTP validation successful for password reset: %s", email)

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logging.Error.Printf("Failed to hash new password for user %s: %v", email, err)
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to hash password",
		)
	}
	logging.Debug.Println("New password hashed successfully")

	// Prepare updates for password and clear OTP
	updates := map[string]interface{}{
		"password":   string(hashedPassword),
		"otp":        "",
		"otp_expiry": time.Time{},
	}

	// Update user record
	if err := s.userRepo.UpdateByFields(&model.User{}, user.ID, updates); err != nil {
		logging.Error.Printf("Failed to update password for user %s: %v", email, err)
		return err
	}
	
	logging.Debug.Printf("Password reset successfully for user: %s", email)
	return nil
}

// GetProfile retrieves user profile by ID
// Parameters:
//   - userID: User's UUID as string
//
// Returns:
//   - *model.User: User profile
//   - error: Error if user not found
func (s *UserAuthService) GetProfile(userID string) (*model.User, error) {
	logging.Debug.Printf("Fetching profile for user ID: %s", userID)

	var user model.User

	// Find user by ID
	if err := s.userRepo.FindById(&user, userID); err != nil {
		logging.Error.Printf("Profile fetch failed - User not found: %s", userID)
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"User not found",
		)
	}
	
	logging.Debug.Printf("Profile fetched successfully for user ID: %s", userID)
	return &user, nil
}

// UpdateProfile updates user's name
// Parameters:
//   - userID: User's UUID as string
//   - name: New name to update
//
// Returns:
//   - *model.User: Updated user profile
//   - error: Error if update fails
func (s *UserAuthService) UpdateProfile(userID string, name string) (*model.User, error) {
	logging.Debug.Printf("Updating profile for user ID: %s, new name: %s", userID, name)

	var user model.User

	// Find user by ID
	if err := s.userRepo.FindById(&user, userID); err != nil {
		logging.Error.Printf("Profile update failed - User not found: %s", userID)
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"User not found",
		)
	}
	logging.Debug.Printf("User found for profile update: %s", userID)

	// Prepare name update
	updates := map[string]interface{}{
		"name": name,
	}

	// Update user record
	if err := s.userRepo.UpdateByFields(&model.User{}, user.ID, updates); err != nil {
		logging.Error.Printf("Failed to update profile for user %s: %v", userID, err)
		return nil, err
	}
	logging.Debug.Printf("Profile updated successfully for user: %s", userID)

	// Reload updated user data
	if err := s.userRepo.FindById(&user, userID); err != nil {
		logging.Error.Printf("Failed to reload updated user %s: %v", userID, err)
		return nil, err
	}

	logging.Debug.Printf("Profile update completed for user ID: %s", userID)
	return &user, nil
}

// ToggleUserBlock toggles the block status of a user
// Parameters:
//   - userID: User's UUID as string
//
// Returns:
//   - *model.User: Updated user with new block status
//   - error: Error if operation fails
func (s *UserAuthService) ToggleUserBlock(userID string) (*model.User, error) {
	logging.Debug.Printf("Toggling block status for user ID: %s", userID)

	var user model.User

	// Find user by ID
	if err := s.userRepo.FindById(&user, userID); err != nil {
		logging.Error.Printf("Toggle block failed - User not found: %s", userID)
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"User not found",
		)
	}
	logging.Debug.Printf("User found, current block status: %v", user.IsBlocked)

	// Toggle block status
	newStatus := !user.IsBlocked
	updates := map[string]interface{}{
		"is_blocked": newStatus,
	}

	// Update user record
	if err := s.userRepo.UpdateByFields(&model.User{}, user.ID, updates); err != nil {
		logging.Error.Printf("Failed to toggle block status for user %s: %v", userID, err)
		return nil, err
	}
	logging.Debug.Printf("Block status toggled to: %v for user: %s", newStatus, userID)

	// Reload updated user data
	if err := s.userRepo.FindById(&user, userID); err != nil {
		logging.Error.Printf("Failed to reload user after block toggle: %s - %v", userID, err)
		return nil, err
	}

	logging.Debug.Printf("User block toggle completed successfully for: %s", userID)
	return &user, nil
}

// GetByID fetches a user by their ID
// Parameters:
//   - userID: User's UUID as string
//
// Returns:
//   - *model.User: User object
//   - error: Error if user not found
func (s *UserAuthService) GetByID(userID string) (*model.User, error) {
	logging.Debug.Printf("Fetching user by ID: %s", userID)

	var user model.User

	if err := s.userRepo.FindById(&user, userID); err != nil {
		logging.Error.Printf("GetByID failed - User not found: %s", userID)
		return nil, apperror.New(
			constant.NOTFOUND,
			"",
			"User not found",
		)
	}

	logging.Debug.Printf("User fetched successfully by ID: %s", userID)
	return &user, nil
}

// GetAllUsers retrieves all users from the database
// Returns:
//   - []model.User: Slice of all users
//   - error: Error if fetch fails
func (s *UserAuthService) GetAllUsers() ([]model.User, error) {
	logging.Debug.Println("Fetching all users from database")

	var users []model.User

	if err := s.userRepo.FindAll(&users); err != nil {
		logging.Error.Printf("Failed to fetch all users: %v", err)
		return nil, apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to fetch users",
		)
	}

	logging.Debug.Printf("Successfully fetched %d users", len(users))
	return users, nil
}

// DeleteUserByID deletes a user by their ID
// Parameters:
//   - userID: User's UUID as string
//
// Returns:
//   - error: Error if deletion fails
func (s *UserAuthService) DeleteUserByID(userID string) error {
	logging.Debug.Printf("Attempting to delete user with ID: %s", userID)

	var user model.User

	// Check if user exists
	if err := s.userRepo.FindById(&user, userID); err != nil {
		logging.Error.Printf("Delete failed - User not found: %s", userID)
		return apperror.New(
			constant.NOTFOUND,
			"",
			"User not found",
		)
	}
	logging.Debug.Printf("User found for deletion: %s", userID)

	// Prevent deleting admin users
	if user.Role == "admin" {
		logging.Error.Printf("Delete failed - Attempt to delete admin user: %s", userID)
		return apperror.New(
			constant.FORBIDDEN,
			"",
			"Admin users cannot be deleted",
		)
	}
	logging.Debug.Println("Admin check passed - user is not admin")

	// Delete user
	if err := s.userRepo.Delete(&model.User{}, userID); err != nil {
		logging.Error.Printf("Failed to delete user %s: %v", userID, err)
		return apperror.New(
			constant.INTERNALSERVERERROR,
			"",
			"Failed to delete user",
		)
	}

	logging.Debug.Printf("User deleted successfully: %s", userID)
	return nil
}