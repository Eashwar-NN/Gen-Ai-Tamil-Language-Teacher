package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lang-portal/backend_go/internal/database"
	"github.com/lang-portal/backend_go/internal/handlers"
	"github.com/lang-portal/backend_go/internal/middleware"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Create a default gin router
	r := gin.New()

	// Add middleware
	r.Use(middleware.Logger())
	r.Use(middleware.RequestID())
	r.Use(gin.Recovery())

	// Create handler
	h := handlers.NewHandler(db)

	// Initialize routes
	initializeRoutes(r, h)

	// Start the server
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func initializeRoutes(r *gin.Engine, h *handlers.Handler) {
	// API group
	api := r.Group("/api")

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Dashboard routes
	api.GET("/dashboard/last_study_session", h.GetLastStudySession)
	api.GET("/dashboard/study_progress", h.GetStudyProgress)
	api.GET("/dashboard/quick-stats", h.GetQuickStats)

	// Words routes with pagination validation
	api.GET("/words", middleware.ValidateQueryParams(&middleware.PaginationParams{}), h.GetWords)
	api.GET("/words/:id", h.GetWord)
	api.POST("/words", middleware.ValidateJSON(&middleware.WordRequest{}), h.CreateWord)

	// Groups routes with pagination validation
	api.GET("/groups", middleware.ValidateQueryParams(&middleware.PaginationParams{}), h.GetGroups)
	api.GET("/groups/:id", h.GetGroup)
	api.POST("/groups", middleware.ValidateJSON(&middleware.GroupRequest{}), h.CreateGroup)
	api.GET("/groups/:id/words", middleware.ValidateQueryParams(&middleware.PaginationParams{}), h.GetGroupWords)
	api.GET("/groups/:id/study_sessions", middleware.ValidateQueryParams(&middleware.PaginationParams{}), h.GetGroupStudySessions)

	// Study activities routes
	api.GET("/study_activities/:id", h.GetStudyActivity)
	api.GET("/study_activities/:id/study_sessions",
		middleware.ValidateQueryParams(&middleware.PaginationParams{}), h.GetStudyActivitySessions)
	api.POST("/study_activities", middleware.ValidateJSON(&middleware.StudyActivityRequest{}), h.CreateStudyActivity)

	// Study sessions routes with validation
	api.GET("/study_sessions", middleware.ValidateQueryParams(&middleware.PaginationParams{}), h.GetStudySessions)
	api.GET("/study_sessions/:id", h.GetStudySession)
	api.GET("/study_sessions/:id/words", middleware.ValidateQueryParams(&middleware.PaginationParams{}), h.GetStudySessionWords)
	api.POST("/study_sessions", middleware.ValidateJSON(&middleware.StudySessionRequest{}), h.CreateStudySession)
	api.POST("/study_sessions/:id/words/:word_id/review",
		middleware.ValidateJSON(new(middleware.WordReviewRequest)), h.ReviewWord)

	// Reset routes
	api.POST("/reset_history", h.ResetHistory)
	api.POST("/full_reset", h.FullReset)
}
