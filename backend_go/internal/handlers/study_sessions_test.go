package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lang-portal/backend_go/internal/models"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StudySessionsTestSuite struct {
	suite.Suite
	router *gin.Engine
	h      *Handler
	db     *models.DB
}

func (suite *StudySessionsTestSuite) SetupTest() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize test database
	suite.db = setupTestDB(suite.T())

	// Insert test data in the correct order
	_, err := suite.db.Exec(`
		-- First, create groups
		INSERT INTO groups (id, name) VALUES 
		(1, 'Basic Greetings'),
		(2, 'Common Phrases');

		-- Then, create words
		INSERT INTO words (id, tamil, romaji, english) VALUES 
		(1, 'வணக்கம்', 'vanakkam', 'hello'),
		(2, 'நன்றி', 'nandri', 'thank you');

		-- Then, create word-group associations
		INSERT INTO words_groups (word_id, group_id) VALUES 
		(1, 1),
		(2, 1);

		-- Then, create study sessions
		INSERT INTO study_sessions (id, group_id) VALUES 
		(1, 1),
		(2, 2);

		-- Then, create study activities
		INSERT INTO study_activities (id, study_session_id, group_id) VALUES 
		(1, 1, 1),
		(2, 2, 2);

		-- Update study sessions with study activities
		UPDATE study_sessions SET study_activities_id = 1 WHERE id = 1;
		UPDATE study_sessions SET study_activities_id = 2 WHERE id = 2;

		-- Finally, create word review items
		INSERT INTO word_review_items (word_id, study_session_id, correct) VALUES
		(1, 1, true),
		(2, 1, false);
	`)
	if err != nil {
		suite.T().Fatalf("Failed to insert test data: %v", err)
	}

	// Create handler with test database
	suite.h = NewHandler(suite.db)

	// Setup router with proper parameter handling
	suite.router = gin.New()
	api := suite.router.Group("/api")
	api.GET("/study_sessions", suite.h.GetStudySessions)
	api.GET("/study_sessions/:id", suite.h.GetStudySession)
	api.GET("/study_sessions/:id/words", suite.h.GetStudySessionWords)
	api.POST("/study_sessions/:id/words/:word_id/review", suite.h.ReviewWord)
}

func (suite *StudySessionsTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.DB.Close()
	}
}

func TestStudySessionsTestSuite(t *testing.T) {
	suite.Run(t, new(StudySessionsTestSuite))
}

func (suite *StudySessionsTestSuite) TestGetStudySessions() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/study_sessions", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response struct {
		Items []struct {
			ID               int    `json:"id"`
			ActivityName     string `json:"activity_name"`
			GroupName        string `json:"group_name"`
			StartTime        string `json:"start_time"`
			EndTime          string `json:"end_time"`
			ReviewItemsCount int    `json:"review_items_count"`
		} `json:"items"`
		Pagination struct {
			CurrentPage  int `json:"current_page"`
			TotalPages   int `json:"total_pages"`
			TotalItems   int `json:"total_items"`
			ItemsPerPage int `json:"items_per_page"`
		} `json:"pagination"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), 1, response.Pagination.CurrentPage)
	assert.Equal(suite.T(), 100, response.Pagination.ItemsPerPage)
	assert.NotEmpty(suite.T(), response.Items)
	assert.Equal(suite.T(), 2, len(response.Items))
}

func (suite *StudySessionsTestSuite) TestGetStudySession() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/study_sessions/1", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response struct {
		ID               int    `json:"id"`
		ActivityName     string `json:"activity_name"`
		GroupName        string `json:"group_name"`
		StartTime        string `json:"start_time"`
		EndTime          string `json:"end_time"`
		ReviewItemsCount int    `json:"review_items_count"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), 1, response.ID)
	assert.Equal(suite.T(), "Basic Greetings", response.GroupName)
	assert.Equal(suite.T(), 2, response.ReviewItemsCount)
}

func (suite *StudySessionsTestSuite) TestGetStudySessionWords() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/study_sessions/1/words", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response struct {
		Items []struct {
			Tamil        string `json:"tamil"`
			Romaji       string `json:"romaji"`
			English      string `json:"english"`
			CorrectCount int    `json:"correct_count"`
			WrongCount   int    `json:"wrong_count"`
		} `json:"items"`
		Pagination struct {
			CurrentPage  int `json:"current_page"`
			TotalPages   int `json:"total_pages"`
			TotalItems   int `json:"total_items"`
			ItemsPerPage int `json:"items_per_page"`
		} `json:"pagination"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), 1, response.Pagination.CurrentPage)
	assert.Equal(suite.T(), 100, response.Pagination.ItemsPerPage)
	assert.NotEmpty(suite.T(), response.Items)
	assert.Equal(suite.T(), 2, len(response.Items))
}

func (suite *StudySessionsTestSuite) TestReviewWord() {
	reqBody := map[string]interface{}{
		"correct": true,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/study_sessions/1/words/1/review", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response struct {
		Success   bool   `json:"success"`
		WordID    int    `json:"word_id"`
		SessionID int    `json:"session_id"`
		Correct   bool   `json:"correct"`
		CreatedAt string `json:"created_at"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), 1, response.WordID)
	assert.Equal(suite.T(), 1, response.SessionID)
	assert.True(suite.T(), response.Correct)
	assert.NotEmpty(suite.T(), response.CreatedAt)
}
