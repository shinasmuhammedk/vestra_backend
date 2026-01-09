package handlers

import (
	"net/http"

	"vestra-ecommerce-backend/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	adminUC *usecase.AdminUsecase
}

func NewAdminHandler(adminUC *usecase.AdminUsecase) *AdminHandler {
	return &AdminHandler{adminUC: adminUC}
}

// Request body for block / unblock
type BlockUserRequest struct {
	Block bool `json:"block"` // true = block, false = unblock
}

// PUT /admin/users/:id/block
func (h *AdminHandler) BlockUser(c *gin.Context) {

	// 1️⃣ Get user ID from URL
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
		})
		return
	}

	// 2️⃣ Read request body
	var req BlockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// 3️⃣ Call usecase
	if err := h.adminUC.SetBlockStatus(userID, req.Block); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 4️⃣ Response message
	message := "user unblocked successfully"
	if req.Block {
		message = "user blocked successfully"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}


// GET /admin/users
func (h *AdminHandler) GetAllUsers(c *gin.Context) {

	users, err := h.adminUC.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}




func (h *AdminHandler) DeleteUser(c *gin.Context) {
    idParam := c.Param("id")
    userID, err := uuid.Parse(idParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
        return
    }

    err = h.adminUC.SoftDeleteUser(userID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "user soft deleted successfully"})
}
