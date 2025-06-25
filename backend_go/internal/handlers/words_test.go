package handlers

import (
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

type WordsTestSuite struct {
	suite.Suite
	router *gin.Engine
	h      *Handler
	db     *models.DB
}

func (suite *WordsTestSuite) SetupTest() {
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
		(2, 'நன்றி', 'nandri', 'thank you'),
		(3, 'காலை வணக்கம்', 'kalai vanakkam', 'good morning');

		-- Then, create word-group associations
		INSERT INTO words_groups (word_id, group_id) VALUES 
		(1, 1),
		(2, 2),
		(3, 1);

		-- Then, create study sessions
		INSERT INTO study_sessions (id, group_id) VALUES (1, 1);

		-- Then, create study activities
		INSERT INTO study_activities (id, study_session_id, group_id) VALUES (1, 1, 1);

		-- Update study session with study activity
		UPDATE study_sessions SET study_activities_id = 1 WHERE id = 1;

		-- Finally, create word review items
		INSERT INTO word_review_items (word_id, study_session_id, correct) VALUES
		(1, 1, true),
		(1, 1, true),
		(1, 1, false);
	`)
	if err != nil {
		suite.T().Fatalf("Failed to insert test data: %v", err)
	}

	// Create handler with test database
	suite.h = NewHandler(suite.db)

	// Setup router with proper parameter handling
	suite.router = gin.New()
	api := suite.router.Group("/api")
	api.GET("/words", suite.h.GetWords)
	api.GET("/words/:id", suite.h.GetWord)
}

func (suite *WordsTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.DB.Close()
	}
}

func TestWordsTestSuite(t *testing.T) {
	suite.Run(t, new(WordsTestSuite))
}

func (suite *WordsTestSuite) TestGetWords() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/words", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response struct {
		Items []struct {
			Tamil   string `json:"tamil"`
			Romaji  string `json:"romaji"`
			English string `json:"english"`
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

	// Check pagination
	assert.Equal(suite.T(), 1, response.Pagination.CurrentPage)
	assert.Equal(suite.T(), 1, response.Pagination.TotalPages)
	assert.Equal(suite.T(), 3, response.Pagination.TotalItems)
	assert.Equal(suite.T(), 100, response.Pagination.ItemsPerPage)

	// Check items
	assert.Len(suite.T(), response.Items, 3)
	assert.Equal(suite.T(), "வணக்கம்", response.Items[0].Tamil)
	assert.Equal(suite.T(), "vanakkam", response.Items[0].Romaji)
	assert.Equal(suite.T(), "hello", response.Items[0].English)
}

func (suite *WordsTestSuite) TestGetWord() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/words/1", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response struct {
		Tamil   string `json:"tamil"`
		Romaji  string `json:"romaji"`
		English string `json:"english"`
		Stats   struct {
			CorrectCount int `json:"correct_count"`
			WrongCount   int `json:"wrong_count"`
		} `json:"stats"`
		Groups []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"groups"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "வணக்கம்", response.Tamil)
	assert.Equal(suite.T(), "vanakkam", response.Romaji)
	assert.Equal(suite.T(), "hello", response.English)
	assert.Equal(suite.T(), 2, response.Stats.CorrectCount)
	assert.Equal(suite.T(), 1, response.Stats.WrongCount)
	assert.Len(suite.T(), response.Groups, 1)
	assert.Equal(suite.T(), 1, response.Groups[0].ID)
	assert.Equal(suite.T(), "Basic Greetings", response.Groups[0].Name)
}

func (suite *WordsTestSuite) TestGetWordNotFound() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/words/999", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Word not found", response["error"])
}
