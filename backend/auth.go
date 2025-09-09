package main

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Секретный ключ для JWT (потом вынесем в .env)
var jwtSecret = []byte("supersecretkey123")

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Структура токена ответа
type TokenResponse struct {
	Token string `json:"token"`
}

// Функция хеширования пароля
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Проверка пароля
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Логин
func LoginHandler(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// TODO: достать пользователя из БД по email
	// для теста — захардкодим
	if req.Email != "admin@synergium.local" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "wrong email"})
	}

	// сравниваем пароль
	if !CheckPasswordHash(req.Password, "$2a$14$SgeH8h4Uq2Pqa0yBN/.3LejxGQOHuHpk7kH3hGpM9XBlIpr7HLDRa") {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "wrong password"})
	}

	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": req.Email,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not create token"})
	}

	return c.JSON(http.StatusOK, TokenResponse{Token: tokenString})

}
