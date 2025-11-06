package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"tracking-api/models"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var JwtKey = []byte("supersecretkey")

// REGISTER
func Register(c *gin.Context, db *gorm.DB) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Println("Error terjadi:", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error terjadi:", err)
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	if err := db.Create(&user).Error; err != nil {
		fmt.Println("Error terjadi:", err)
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": fmt.Sprintf("%d", user.ID),
		"email":   user.Email,
		"role":    user.Role,
	})

	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		fmt.Println("Error terjadi:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(200, gin.H{
		"message": "User registered",
		"token":   tokenString,
		"user":    user,
	})
}

// LOGIN
func Login(c *gin.Context, db *gorm.DB) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var user models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Println("Error terjadi:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("User %s is %s years old.", input.Email, input.Password)
	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		fmt.Println("Error terjadi:", err)
		return
	}

	pass := strings.TrimSpace(input.Password)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		fmt.Println("Error terjadi:", err)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": fmt.Sprintf("%d", user.ID),
		"email":   user.Email,
		"role":    user.Role,
	})

	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login success",
		"token":   tokenString,
	})
}

func LoginAdmin(c *gin.Context, db *gorm.DB) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var user models.User

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin can login"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": fmt.Sprintf("%d", user.ID),
		"email":   user.Email,
		"role":    user.Role,
	})

	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login success",
		"token":   tokenString,
	})
}
