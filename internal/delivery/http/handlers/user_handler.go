package handlers

import (
	"fmt"
	"net/http"

	"vestra-ecommerce-backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUseCase *usecase.UserUseCase
}

func NewUserHandler(userUC *usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUseCase: userUC}
}

// GET /user/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id") // from middleware

    
	// 🔴 DEBUG
	fmt.Println("USER ID FROM TOKEN:", userID)
    
    
	user, err := h.userUseCase.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}



// =====================
// UPDATE PROFILE
// =====================

type UpdateProfileRequest struct {
	Name string `json:"name" binding:"required,min=2"`
}

// PUT /user/profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id") // from AuthMiddleware

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userUseCase.UpdateProfile(userID, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}