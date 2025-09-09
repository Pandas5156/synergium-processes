package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func main() {
	// Читаем конфиг из окружения
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// Подключаем пул к базе
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}
	log.Println("Connected to PostgreSQL!")

	// Создаём Echo
	e := echo.New()
	e.POST("/auth/login", LoginHandler)

	// Эндпоинт для проверки приложения
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	// Эндпоинт для проверки БД
	e.GET("/dbping", func(c echo.Context) error {
		if err := pool.Ping(context.Background()); err != nil {
			return c.String(http.StatusInternalServerError, "db error")
		}
		return c.String(http.StatusOK, "db pong")
	})

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
