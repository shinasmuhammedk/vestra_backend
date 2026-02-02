package controller

import (
	"github.com/gofiber/fiber/v2"

	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/jwt"
	"vestra-ecommerce/utils/logging"
	"vestra-ecommerce/utils/response"
	"vestra-ecommerce/utils/utils/apperror"
)

type UserAuthController struct {
	authService *services.UserAuthService
	jwtManager  *jwt.JWTManager
}

func NewUserAuthController(service *services.UserAuthService, manager *jwt.JWTManager) *UserAuthController {
	logging.Debug.Println("UserAuthController initialized")
	return &UserAuthController{
		authService: service,
		jwtManager:  manager,
	}
}

// Request structs
type signupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type verifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

type updateProfileRequest struct {
	Name string `json:"name"`
}

// Signup handles user registration
func (c *UserAuthController) Signup(ctx *fiber.Ctx) error {
	logging.Debug.Println("Signup endpoint called")

	var req signupRequest
	if err := ctx.BodyParser(&req); err != nil {
		logging.Error.Printf("Signup - Invalid request body: %v", err)
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Signup attempt for email: %s", req.Email)
	
	if err := c.authService.Signup(req.Name, req.Email, req.Password); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("Signup failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("Signup error: %v", err)
		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Something went wrong",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Signup successful - OTP sent to: %s", req.Email)
	return response.Success(
		ctx,
		constant.CREATED,
		"OTP sent to your email",
		"",
		nil,
	)
}

// VerifyOTP verifies user's OTP
func (c *UserAuthController) VerifyOTP(ctx *fiber.Ctx) error {
	logging.Debug.Println("VerifyOTP endpoint called")

	var req verifyOTPRequest
	if err := ctx.BodyParser(&req); err != nil {
		logging.Error.Printf("VerifyOTP - Invalid request body: %v", err)
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Verifying OTP for email: %s", req.Email)
	
	if err := c.authService.VerifyOTP(req.Email, req.OTP); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("VerifyOTP failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("VerifyOTP error: %v", err)
		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Something went wrong",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("OTP verified successfully for: %s", req.Email)
	return response.Success(
		ctx,
		constant.SUCCESS,
		"Account verified successfully",
		"",
		nil,
	)
}

// Login authenticates user
func (c *UserAuthController) Login(ctx *fiber.Ctx) error {
	logging.Debug.Println("Login endpoint called")

	var req loginRequest
	if err := ctx.BodyParser(&req); err != nil {
		logging.Error.Printf("Login - Invalid request body: %v", err)
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"Invalid request payload",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Login attempt for email: %s", req.Email)

	user, err := c.authService.Login(req.Email, req.Password)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("Login failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("Login error: %v", err)
		return response.Error(
			ctx,
			constant.UNAUTHORIZED,
			"Invalid credentials",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Login successful - generating tokens for user: %s", user.ID)

	accessToken, err := c.jwtManager.GenerateAccessToken(user.ID.String(), user.Role)
	if err != nil {
		logging.Error.Printf("Failed to generate access token: %v", err)
		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Failed to generate access token",
			"",
			err.Error(),
		)
	}

	refreshToken, err := c.jwtManager.GenerateRefreshToken(user.ID.String(), user.Role)
	if err != nil {
		logging.Error.Printf("Failed to generate refresh token: %v", err)
		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Failed to generate refresh token",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Tokens generated successfully for user: %s", user.ID)
	return response.Success(
		ctx,
		constant.SUCCESS,
		"Login successful",
		"",
		fiber.Map{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	)
}

// RefreshToken generates new access token
func (c *UserAuthController) RefreshToken(ctx *fiber.Ctx) error {
	logging.Debug.Println("RefreshToken endpoint called")

	var req refreshRequest
	if err := ctx.BodyParser(&req); err != nil {
		logging.Error.Printf("RefreshToken - Invalid request body: %v", err)
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"Invalid request payload",
			"",
			nil,
		)
	}

	claims, err := c.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		logging.Error.Printf("Invalid refresh token: %v", err)
		return response.Error(
			ctx,
			constant.UNAUTHORIZED,
			"Invalid or expired refresh token",
			"",
			nil,
		)
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		logging.Error.Println("Invalid token claims - missing user_id")
		return response.Error(
			ctx,
			constant.UNAUTHORIZED,
			"Invalid token claims",
			"",
			nil,
		)
	}
    
	role, ok := claims["role"].(string)
	if !ok || role == "" {
		role = "user"
		logging.Debug.Println("Role not found in token, defaulting to 'user'")
	}

	accessToken, err := c.jwtManager.GenerateAccessToken(userID, role)
	if err != nil {
		logging.Error.Printf("Failed to generate access token: %v", err)
		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Failed to generate access token",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Access token refreshed for user: %s", userID)
	return response.Success(
		ctx,
		constant.SUCCESS,
		"Access token refreshed",
		"",
		fiber.Map{
			"access_token": accessToken,
		},
	)
}

// ForgotPassword initiates password reset
func (c *UserAuthController) ForgotPassword(ctx *fiber.Ctx) error {
	logging.Debug.Println("ForgotPassword endpoint called")

	var req forgotPasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		logging.Error.Printf("ForgotPassword - Invalid request body: %v", err)
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Forgot password request for email: %s", req.Email)
	
	if err := c.authService.ForgotPassword(req.Email); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("ForgotPassword failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		logging.Error.Printf("ForgotPassword error: %v", err)
		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Something went wrong",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Password reset OTP processed for: %s", req.Email)
	return response.Success(
		ctx,
		constant.SUCCESS,
		"If email exists, OTP sent to the inbox",
		"",
		nil,
	)
}

// ResetPassword resets user's password
func (c *UserAuthController) ResetPassword(ctx *fiber.Ctx) error {
	logging.Debug.Println("ResetPassword endpoint called")

	var req resetPasswordRequest

	if err := ctx.BodyParser(&req); err != nil {
		logging.Error.Printf("ResetPassword - Invalid request body: %v", err)
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		logging.Error.Println("ResetPassword - missing required fields")
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"All fields are required",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Resetting password for email: %s", req.Email)
	
	if err := c.authService.ResetPassword(
		req.Email,
		req.OTP,
		req.NewPassword,
	); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("ResetPassword failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		logging.Error.Printf("ResetPassword error: %v", err)
		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Something went wrong",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Password reset successful for: %s", req.Email)
	return response.Success(
		ctx,
		constant.SUCCESS,
		"Password reset successfully",
		"",
		nil,
	)
}

// GetProfile fetches user profile
func (c *UserAuthController) GetProfile(ctx *fiber.Ctx) error {
	logging.Debug.Println("GetProfile endpoint called")

	userID := ctx.Locals("user_id")
	if userID == nil {
		logging.Error.Println("GetProfile - User ID not found in context")
		return response.Error(
			ctx,
			constant.UNAUTHORIZED,
			"Unauthorized",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Fetching profile for user: %s", userID)
	
	user, err := c.authService.GetProfile(userID.(string))
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("GetProfile failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		logging.Error.Printf("GetProfile error: %v", err)
		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Something went wrong",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Profile fetched successfully for user: %s", userID)
	return response.Success(
		ctx,
		constant.SUCCESS,
		"Profile fetched successfully",
		"",
		user,
	)
}

// UpdateProfile updates user's name
func (c *UserAuthController) UpdateProfile(ctx *fiber.Ctx) error {
	logging.Debug.Println("UpdateProfile endpoint called")

	var req updateProfileRequest
	if err := ctx.BodyParser(&req); err != nil {
		logging.Error.Printf("UpdateProfile - Invalid request body: %v", err)
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"Invalid request body",
			"",
			nil,
		)
	}

	if req.Name == "" {
		logging.Error.Println("UpdateProfile - Name is required")
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"Name is required",
			"",
			nil,
		)
	}

	userID := ctx.Locals("user_id")
	if userID == nil {
		logging.Error.Println("UpdateProfile - User ID not found in context")
		return response.Error(
			ctx,
			constant.UNAUTHORIZED,
			"Unauthorized",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Updating profile for user: %s, new name: %s", userID, req.Name)
	
	user, err := c.authService.UpdateProfile(userID.(string), req.Name)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("UpdateProfile failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("UpdateProfile error: %v", err)
		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Something went wrong",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Profile updated successfully for user: %s", userID)
	return response.Success(
		ctx,
		constant.SUCCESS,
		"Profile updated successfully",
		"",
		user,
	)
}

// ToggleUserBlock toggles user block status (admin only)
func (c *UserAuthController) ToggleUserBlock(ctx *fiber.Ctx) error {
	logging.Debug.Println("ToggleUserBlock endpoint called")

	currentUserID := ctx.Locals("user_id")
	if currentUserID == nil {
		logging.Error.Println("ToggleUserBlock - User ID not found in context")
		return response.Error(
			ctx,
			constant.UNAUTHORIZED,
			"Unauthorized",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Admin user checking permissions: %s", currentUserID)
	
	currentUser, err := c.authService.GetByID(currentUserID.(string))
	if err != nil {
		logging.Error.Printf("Failed to get current user: %v", err)
		return response.Error(
			ctx,
			constant.UNAUTHORIZED,
			"Unauthorized",
			"",
			nil,
		)
	}
	
	if currentUser.Role != "admin" {
		logging.Error.Printf("Non-admin user attempted to toggle block: %s", currentUserID)
		return response.Error(
			ctx,
			constant.FORBIDDEN,
			"Admin access required",
			"",
			nil,
		)
	}

	targetID := ctx.Params("id")
	if targetID == "" {
		logging.Error.Println("ToggleUserBlock - Target user ID required")
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"User ID required",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Admin %s toggling block for user: %s", currentUserID, targetID)
	
	updatedUser, err := c.authService.ToggleUserBlock(targetID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("ToggleUserBlock failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		logging.Error.Printf("ToggleUserBlock error: %v", err)
		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Something went wrong",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("User block status toggled - User: %s, New status: %v", targetID, updatedUser.IsBlocked)
	return response.Success(
		ctx,
		constant.SUCCESS,
		"User block status toggled",
		"",
		updatedUser,
	)
}

// GetAllUsers fetches all users (admin only)
func (c *UserAuthController) GetAllUsers(ctx *fiber.Ctx) error {
	logging.Debug.Println("GetAllUsers endpoint called")

	users, err := c.authService.GetAllUsers()
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("GetAllUsers failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(
				ctx,
				appErr.Status,
				appErr.Message,
				appErr.Code,
				nil,
			)
		}

		logging.Error.Printf("GetAllUsers error: %v", err)
		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Failed to fetch users",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("Fetched %d users", len(users))
	return response.Success(
		ctx,
		constant.SUCCESS,
		"Users fetched successfully",
		"",
		users,
	)
}

// DeleteUserByID deletes user by ID (admin only)
func (c *UserAuthController) DeleteUserByID(ctx *fiber.Ctx) error {
	logging.Debug.Println("DeleteUserByID endpoint called")

	targetID := ctx.Params("id")
	if targetID == "" {
		logging.Error.Println("DeleteUserByID - User ID is required")
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"User ID is required",
			"",
			nil,
		)
	}

	currentUserID := ctx.Locals("user_id").(string)
	if currentUserID == targetID {
		logging.Error.Printf("Admin attempted to delete own account: %s", currentUserID)
		return response.Error(
			ctx,
			constant.FORBIDDEN,
			"Admin cannot delete own account",
			"",
			nil,
		)
	}

	logging.Debug.Printf("Admin %s deleting user: %s", currentUserID, targetID)
	
	if err := c.authService.DeleteUserByID(targetID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			logging.Error.Printf("DeleteUserByID failed: %s - %s", appErr.Code, appErr.Message)
			return response.Error(
				ctx,
				appErr.Status,
				appErr.Message,
				appErr.Code,
				nil,
			)
		}

		logging.Error.Printf("DeleteUserByID error: %v", err)
		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Failed to delete user",
			"",
			err.Error(),
		)
	}

	logging.Debug.Printf("User deleted successfully: %s", targetID)
	return response.Success(
		ctx,
		constant.SUCCESS,
		"User deleted successfully",
		"",
		nil,
	)
}