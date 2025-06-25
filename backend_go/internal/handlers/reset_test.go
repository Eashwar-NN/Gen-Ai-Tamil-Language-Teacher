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

type ResetTestSuite struct {
	suite.Suite
	router *gin.Engine
	h      *Handler
	db     *models.DB
}

func (suite *ResetTestSuite) SetupTest() {
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
		suite.T().Fatalf("Failed to insert test data: %v", err)
	}

	// Create handler with test database
	suite.h = NewHandler(suite.db)

	// Setup router with proper parameter handling
	suite.router = gin.New()
	api := suite.router.Group("/api")
	api.POST("/reset_history", suite.h.ResetHistory)
	api.POST("/full_reset", suite.h.FullReset)
}

func (suite *ResetTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.DB.Close()
	}
}

func TestResetTestSuite(t *testing.T) {
	suite.Run(t, new(ResetTestSuite))
}

func (suite *ResetTestSuite) TestResetHistory() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/reset_history", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response struct {
		Success        bool   `json:"success"`
		Message        string `json:"message"`
		ResetTimestamp string `json:"reset_timestamp"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "Study history has been reset", response.Message)
	assert.NotEmpty(suite.T(), response.ResetTimestamp)

	// Verify study history is cleared
	var count int
	err = suite.db.QueryRow("SELECT COUNT(*) FROM word_review_items").Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)

	err = suite.db.QueryRow("SELECT COUNT(*) FROM study_sessions").Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)

	err = suite.db.QueryRow("SELECT COUNT(*) FROM study_activities").Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)

	// Verify words and groups are not cleared
	err = suite.db.QueryRow("SELECT COUNT(*) FROM words").Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, count)

	err = suite.db.QueryRow("SELECT COUNT(*) FROM groups").Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, count)
}

func (suite *ResetTestSuite) TestFullReset() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/full_reset", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response struct {
		Success        bool   `json:"success"`
		Message        string `json:"message"`
		ResetTimestamp string `json:"reset_timestamp"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.True(suite.T(), response.Success)
	assert.Equal(suite.T(), "System has been fully reset", response.Message)
	assert.NotEmpty(suite.T(), response.ResetTimestamp)

	// Verify all tables are cleared
	var count int
	err = suite.db.QueryRow("SELECT COUNT(*) FROM word_review_items").Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)

	err = suite.db.QueryRow("SELECT COUNT(*) FROM study_sessions").Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)

	err = suite.db.QueryRow("SELECT COUNT(*) FROM study_activities").Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)

	err = suite.db.QueryRow("SELECT COUNT(*) FROM words").Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)

	err = suite.db.QueryRow("SELECT COUNT(*) FROM groups").Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)

	err = suite.db.QueryRow("SELECT COUNT(*) FROM words_groups").Scan(&count)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)
}
