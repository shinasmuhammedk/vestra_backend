package controller

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"vestra-ecommerce/src/services"
	constant "vestra-ecommerce/utils/constants"
	"vestra-ecommerce/utils/jwt"
	"vestra-ecommerce/utils/response"
	"vestra-ecommerce/utils/utils/apperror"
)

type UserAuthController struct {
	authService *services.UserAuthService
	jwtManager  *jwt.JWTManager
}

func NewUserAuthController(service *services.UserAuthService, manager *jwt.JWTManager) *UserAuthController {
	return &UserAuthController{
		authService: service,
		jwtManager:  manager,
	}
}

// ------------------ Signup ------------------

type signupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *UserAuthController) Signup(ctx *fiber.Ctx) error {
	var req signupRequest
	if err := ctx.BodyParser(&req); err != nil {
		log.Println("Signup BodyParser error:", err)
		return response.Error(ctx, constant.BADREQUEST, "Invalid request body", "INVALID_REQUEST", nil)
	}

	if err := c.authService.Signup(req.Name, req.Email, req.Password); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(ctx, constant.INTERNALSERVERERROR, "Something went wrong", "INTERNAL_ERROR", err.Error())
	}

	return response.Success(ctx, constant.CREATED, "OTP sent to your email", "AUTH_OTP_SENT", nil)
}

// ------------------ Verify OTP ------------------

type verifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func (c *UserAuthController) VerifyOTP(ctx *fiber.Ctx) error {
	var req verifyOTPRequest
	if err := ctx.BodyParser(&req); err != nil {
		log.Println("VerifyOTP BodyParser error:", err)
		return response.Error(ctx, constant.BADREQUEST, "Invalid request body", "INVALID_REQUEST", nil)
	}

	if err := c.authService.VerifyOTP(req.Email, req.OTP); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(ctx, constant.INTERNALSERVERERROR, "Something went wrong", "INTERNAL_ERROR", err.Error())
	}

	return response.Success(ctx, constant.SUCCESS, "Account verified successfully", "ACCOUNT_VERIFIED", nil)
}

// ------------------ Login ------------------

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (c *UserAuthController) Login(ctx *fiber.Ctx) error {
	var req loginRequest
	if err := ctx.BodyParser(&req); err != nil {
		log.Println("Login BodyParser error:", err)
		return response.Error(ctx, constant.BADREQUEST, "Invalid request payload", "INVALID_REQUEST", nil)
	}

	log.Println("Login request received:", req.Email)

	user, err := c.authService.Login(req.Email, req.Password)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(ctx, constant.UNAUTHORIZED, "Invalid credentials", "LOGIN_FAILED", nil)
	}

	accessToken, err := c.jwtManager.GenerateAccessToken(user.ID.String())
	if err != nil {
		return response.Error(ctx, constant.INTERNALSERVERERROR, "Failed to generate access token", "TOKEN_GENERATION_FAILED", err.Error())
	}

	refreshToken, err := c.jwtManager.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return response.Error(ctx, constant.INTERNALSERVERERROR, "Failed to generate refresh token", "TOKEN_GENERATION_FAILED", err.Error())
	}

	return response.Success(ctx, constant.SUCCESS, "Login successful", "LOGIN_SUCCESS", fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// ------------------ Refresh Token ------------------

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (c *UserAuthController) RefreshToken(ctx *fiber.Ctx) error {
	var req refreshRequest
	if err := ctx.BodyParser(&req); err != nil {
		log.Println("RefreshToken BodyParser error:", err)
		return response.Error(ctx, constant.BADREQUEST, "Invalid request payload", "INVALID_REQUEST", nil)
	}

	claims, err := c.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return response.Error(ctx, constant.UNAUTHORIZED, "Invalid or expired refresh token", "INVALID_REFRESH_TOKEN", nil)
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return response.Error(ctx, constant.UNAUTHORIZED, "Invalid token claims", "INVALID_REFRESH_TOKEN", nil)
	}

	accessToken, err := c.jwtManager.GenerateAccessToken(userID)
	if err != nil {
		return response.Error(ctx, constant.INTERNALSERVERERROR, "Failed to generate access token", "TOKEN_GENERATION_FAILED", err.Error())
	}

	return response.Success(ctx, constant.SUCCESS, "Access token refreshed", "ACCESS_TOKEN_REFRESHED", fiber.Map{
		"access_token": accessToken,
	})
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (c *UserAuthController) ForgotPassword(ctx *fiber.Ctx) error {
	var req forgotPasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"Invalid request body",
			"INVALID_REQUEST",
			nil,
		)
	}

	if err := c.authService.ForgotPassword(req.Email); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Something went wrong",
			"INTERNAL_ERROR",
			err.Error(),
		)
	}

	return response.Success(
		ctx,
		constant.SUCCESS,
		"If email exists, OTP sent to the inbox",
		"FORGOT_PASSWORD_EMAIL_SENT",
		nil,
	)
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

func (c *UserAuthController) ResetPassword(ctx *fiber.Ctx) error {
	var req resetPasswordRequest

	if err := ctx.BodyParser(&req); err != nil {
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"Invalid request body",
			"INVALID_REQUEST",
			nil,
		)
	}

	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"All fields are required",
			"FIELDS_REQUIRED",
			nil,
		)
	}

	if err := c.authService.ResetPassword(
		req.Email,
		req.OTP,
		req.NewPassword,
	); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Something went wrong",
			"INTERNAL_ERROR",
			err.Error(),
		)
	}

	return response.Success(
		ctx,
		constant.SUCCESS,
		"Password reset successfully",
		"PASSWORD_RESET_SUCCESS",
		nil,
	)
}

func (c *UserAuthController) GetProfile(ctx *fiber.Ctx) error {
	// Get user_id from middleware
	userID := ctx.Locals("user_id")
	if userID == nil {
		return response.Error(
			ctx,
			constant.UNAUTHORIZED,
			"Unauthorized",
			"UNAUTHORIZED",
			nil,
		)
	}

	// Call service to fetch user
	user, err := c.authService.GetProfile(userID.(string))
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}

		return response.Error(
			ctx,
			constant.INTERNALSERVERERROR,
			"Something went wrong",
			"INTERNAL_ERROR",
			err.Error(),
		)
	}

	return response.Success(
		ctx,
		constant.SUCCESS,
		"Profile fetched successfully",
		"PROFILE_FETCHED",
		user,
	)
}

type updateProfileRequest struct {
	Name string `json:"name"` // only name is allowed
}

func (c *UserAuthController) UpdateProfile(ctx *fiber.Ctx) error {
	var req updateProfileRequest
	if err := ctx.BodyParser(&req); err != nil {
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"Invalid request body",
			"INVALID_REQUEST",
			nil,
		)
	}

	if req.Name == "" {
		return response.Error(
			ctx,
			constant.BADREQUEST,
			"Name is required",
			"INVALID_REQUEST",
			nil,
		)
	}

	userID := ctx.Locals("user_id")
	if userID == nil {
		return response.Error(
			ctx,
			constant.UNAUTHORIZED,
			"Unauthorized",
			"UNAUTHORIZED",
			nil,
		)
	}

	user, err := c.authService.UpdateProfile(userID.(string), req.Name)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(ctx, constant.INTERNALSERVERERROR, "Something went wrong", "INTERNAL_ERROR", err.Error())
	}

	return response.Success(
		ctx,
		constant.SUCCESS,
		"Profile updated successfully",
		"PROFILE_UPDATED",
		user,
	)
}




func (c *UserAuthController) ToggleUserBlock(ctx *fiber.Ctx) error {
	// 1️⃣ Get current user ID from JWT
	currentUserID := ctx.Locals("user_id")
	if currentUserID == nil {
		return response.Error(ctx, constant.UNAUTHORIZED, "Unauthorized", "UNAUTHORIZED", nil)
	}

	// 2️⃣ Check if current user is ADMIN
	currentUser, err := c.authService.GetByID(currentUserID.(string))
	if err != nil {
		return response.Error(ctx, constant.UNAUTHORIZED, "Unauthorized", "UNAUTHORIZED", nil)
	}
	if currentUser.Role != "admin" {
		return response.Error(ctx, constant.FORBIDDEN, "Admin access required", "FORBIDDEN", nil)
	}

	// 3️⃣ Get target user ID from URL
	targetID := ctx.Params("id")
	if targetID == "" {
		return response.Error(ctx, constant.BADREQUEST, "User ID required", "INVALID_REQUEST", nil)
	}

	// 4️⃣ Call service to toggle is_blocked
	updatedUser, err := c.authService.ToggleUserBlock(targetID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.Error(ctx, appErr.Status, appErr.Message, appErr.Code, nil)
		}
		return response.Error(ctx, constant.INTERNALSERVERERROR, "Something went wrong", "INTERNAL_ERROR", err.Error())
	}

	return response.Success(ctx, constant.SUCCESS, "User block status toggled", "USER_BLOCK_TOGGLED", updatedUser)
}
