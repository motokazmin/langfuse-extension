package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// TODO: Это базовая структура проекта
// Намеренно НЕ полностью реализована - вы будете дорабатывать
// её с помощью Cursor AI, тестируя разные промпты

func main() {
	router := gin.Default()

	// CORS для разработки
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// TODO: Добавить роуты для:
	// - POST /auth/register
	// - POST /auth/login
	// - GET /posts
	// - POST /posts (требует авторизацию)
	// - GET /posts/:id
	// - PUT /posts/:id (требует авторизацию)
	// - DELETE /posts/:id (требует авторизацию)
	// - POST /posts/:id/comments

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	log.Println("🚀 Blog API запущен на :8080")
	router.Run(":8080")
}
