package main

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Секретный ключ для JWT (в будущем — хранить в .env)
var jwtSecret = []byte("supersecretkey123")

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name,omitempty"`
	Role     string `json:"role,omitempty"` // optional: admin, supervisor, operator, curator, decision_maker
}

type TokenResponse struct {
	Token string `json:"token"`
}

// Хеширование пароля
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12) // cost 12 — быстрее, безопасно
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// RegisterHandler — создает пользователя в БД
// Внимание: пока открыт, используйте для первого создания admin, потом можно закрыть / защитить.
func RegisterHandler(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "email and password required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверим, нет ли уже пользователя с таким email
	var exists bool
	err := dbPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", req.Email).Scan(&exists)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}
	if exists {
		return c.JSON(http.StatusConflict, map[string]string{"error": "user already exists"})
	}

	// Получим role_id по имени роли (если роль не указана — "operator")
	roleName := req.Role
	if roleName == "" {
		roleName = "operator"
	}
	var roleID int
	err = dbPool.QueryRow(ctx, "SELECT id FROM roles WHERE name=$1", roleName).Scan(&roleID)
	if err != nil {
		// если роли нет, вернём ошибку
		if err == pgx.ErrNoRows {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "role not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}

	// Хешируем пароль
	hash, err := HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not hash password"})
	}

	// Вставляем пользователя
	var newID string
	err = dbPool.QueryRow(ctx,
		"INSERT INTO users (email, password_hash, full_name, role_id) VALUES ($1, $2, $3, $4) RETURNING id",
		req.Email, hash, req.FullName, roleID).Scan(&newID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not create user"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"id": newID})
}

// LoginHandler — логин через БД
func LoginHandler(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "email and password required"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id string
	var email string
	var passwordHash string
	var roleName string

	err := dbPool.QueryRow(ctx,
		"SELECT u.id, u.email, u.password_hash, coalesce(r.name, '') FROM users u LEFT JOIN roles r ON u.role_id = r.id WHERE u.email=$1",
		req.Email).Scan(&id, &email, &passwordHash, &roleName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
	}

	// Проверка пароля
	if !CheckPasswordHash(req.Password, passwordHash) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	// Формируем JWT (claims: sub=id, email, role, exp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   id,
		"email": email,
		"role":  roleName,
		"exp":   time.Now().Add(72 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not create token"})
	}

	return c.JSON(http.StatusOK, TokenResponse{Token: tokenString})
}
