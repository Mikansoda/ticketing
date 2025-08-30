package controller

import (
	"net/http"

	"ticketing/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service service.AuthService
}

func NewAuthController(s service.AuthService) *AuthController {
	return &AuthController{service: s}
}

// Struct request for endpoint register, verify, login, and refresh
type registerReq struct {
    FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=8,max=20,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type verifyReq struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register
func (a *AuthController) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid input data",
			"detail":  err.Error(),
		})
		return
	}

    if err := a.service.Register(c, req.FullName, req.Username, req.Email, req.Password, "user"); err != nil {		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to register, try again",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registered successfully, check your email for OTP",
	})
}

// Verify OTP
func (a *AuthController) VerifyOTP(c *gin.Context) {
	var req verifyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid input data",
			"detail":  err.Error(),
		})
		return
	}

	if err := a.service.VerifyOTP(c, req.Email, req.OTP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to verify OTP",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account successfully verified",
	})
}

// Login
func (a *AuthController) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid input data",
			"detail":  err.Error(),
		})
		return
	}

	access, refresh, err := a.service.Login(c, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Login failed",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

// Refresh token
func (a *AuthController) Refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid input data",
			"detail":  err.Error(),
		})
		return
	}

	access, refresh, err := a.service.Refresh(c, req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Failed to refresh token",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

// Log out
func (a *AuthController) Logout(c *gin.Context) {
	bearer := c.GetHeader("Authorization")
	var token string
	if len(bearer) > 7 && bearer[:7] == "Bearer " {
		token = bearer[7:]
	}
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Missing access token",
		})
		return
	}

	if err := a.service.Logout(c, token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed to logout",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

func (a *AuthController) Profile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"userID": c.GetString("userID"),
		"email":  c.GetString("email"),
		"role":   c.GetString("role"),
	})
}

func (a *AuthController) AdminDashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to admin dashboard",
	})
}
