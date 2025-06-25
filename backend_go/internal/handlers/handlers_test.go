package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lang-portal/backend_go/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *models.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Enable foreign key support
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Run migrations
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS words (
			id INTEGER PRIMARY KEY,
			tamil TEXT NOT NULL,
			romaji TEXT NOT NULL,
			english TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS groups (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS words_groups (
			id INTEGER PRIMARY KEY,
			word_id INTEGER,
			group_id INTEGER,
			FOREIGN KEY (word_id) REFERENCES words(id),
			FOREIGN KEY (group_id) REFERENCES groups(id)
		);

		CREATE TABLE IF NOT EXISTS study_sessions (
			id INTEGER PRIMARY KEY,
			group_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			study_activities_id INTEGER,
			FOREIGN KEY (group_id) REFERENCES groups(id)
		);

		CREATE TABLE IF NOT EXISTS study_activities (
			id INTEGER PRIMARY KEY,
			study_session_id INTEGER,
			group_id INTEGER,
			name TEXT NOT NULL DEFAULT 'Vocabulary Quiz',
			thumbnail_url TEXT NOT NULL DEFAULT 'https://example.com/thumbnail.jpg',
			description TEXT NOT NULL DEFAULT 'Practice your vocabulary with flashcards',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (group_id) REFERENCES groups(id),
			FOREIGN KEY (study_session_id) REFERENCES study_sessions(id)
		);

		CREATE TABLE IF NOT EXISTS word_review_items (
			id INTEGER PRIMARY KEY,
			word_id INTEGER,
			study_session_id INTEGER,
			correct BOOLEAN,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (word_id) REFERENCES words(id),
			FOREIGN KEY (study_session_id) REFERENCES study_sessions(id)
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	return &models.DB{DB: db}
}

func setupRouter(db *models.DB) (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(db)

	// Setup routes
	r.GET("/api/dashboard/quick-stats", h.GetQuickStats)
	r.GET("/api/words", h.GetWords)
	r.POST("/api/study_sessions/:id/words/:word_id/review", h.ReviewWord)
	r.POST("/api/reset_history", h.ResetHistory)

	return r, h
}

func TestGetQuickStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.DB.Close()
	router, _ := setupRouter(db)

	// Insert test data in the correct order
	_, err := db.Exec(`
		-- First, create groups
		INSERT INTO groups (id, name) VALUES (1, 'Basic Greetings');

		-- Then, create words
		INSERT INTO words (id, tamil, romaji, english) VALUES 
		(1, 'வணக்கம்', 'vanakkam', 'hello'),
		(2, 'நன்றி', 'nandri', 'thank you');

		-- Then, create study sessions
		INSERT INTO study_sessions (id, group_id) VALUES (1, 1);

		-- Then, create study activities
		INSERT INTO study_activities (id, study_session_id, group_id) VALUES (1, 1, 1);

		-- Update study session with study activity
		UPDATE study_sessions SET study_activities_id = 1 WHERE id = 1;

		-- Finally, create word review items
		INSERT INTO word_review_items (word_id, study_session_id, correct) VALUES 
		(1, 1, true),
		(2, 1, false);
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/dashboard/quick-stats", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response struct {
		TotalWords         int     `json:"total_words"`
		WordsStudied       int     `json:"words_studied"`
		TotalStudySessions int     `json:"total_study_sessions"`
		AverageAccuracy    float64 `json:"average_accuracy"`
		StudyStreakDays    int     `json:"study_streak_days"`
	}

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.TotalWords != 2 {
		t.Errorf("Expected 2 total words, got %d", response.TotalWords)
	}
	if response.WordsStudied != 2 {
		t.Errorf("Expected 2 words studied, got %d", response.WordsStudied)
	}
	if response.TotalStudySessions != 1 {
		t.Errorf("Expected 1 study session, got %d", response.TotalStudySessions)
	}
	if response.AverageAccuracy != 50.0 {
		t.Errorf("Expected 50.0%% average accuracy, got %.1f%%", response.AverageAccuracy)
	}
}

func TestReviewWord(t *testing.T) {
	db := setupTestDB(t)
	defer db.DB.Close()
	router, _ := setupRouter(db)

	// Insert test data in the correct order
	_, err := db.Exec(`
		-- First, create groups
		INSERT INTO groups (id, name) VALUES (1, 'Basic Greetings');

		-- Then, create words
		INSERT INTO words (id, tamil, romaji, english) VALUES 
		(1, 'வணக்கம்', 'vanakkam', 'hello');

		-- Then, create study sessions
		INSERT INTO study_sessions (id, group_id) VALUES (1, 1);

		-- Then, create study activities
		INSERT INTO study_activities (id, study_session_id, group_id) VALUES (1, 1, 1);

		-- Update study session with study activity
		UPDATE study_sessions SET study_activities_id = 1 WHERE id = 1;
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/study_sessions/1/words/1/review",
		strings.NewReader(`{"correct": true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response struct {
		Success bool `json:"success"`
		WordID  int  `json:"word_id"`
		Correct bool `json:"correct"`
	}

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success to be true")
	}
	if response.WordID != 1 {
		t.Errorf("Expected word_id 1, got %d", response.WordID)
	}
	if !response.Correct {
		t.Error("Expected correct to be true")
	}
}
