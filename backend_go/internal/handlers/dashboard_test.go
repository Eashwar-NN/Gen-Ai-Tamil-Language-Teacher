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

type DashboardTestSuite struct {
	suite.Suite
	router *gin.Engine
	h      *Handler
	db     *models.DB
}

func (suite *DashboardTestSuite) SetupTest() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize test database
	suite.db = setupTestDB(suite.T())

	// Insert test data
	_, err := suite.db.Exec(`
		INSERT INTO words (tamil, romaji, english) VALUES 
		('வணக்கம்', 'vanakkam', 'hello'),
		('நன்றி', 'nandri', 'thank you');

		INSERT INTO groups (name) VALUES ('Basic Greetings');

		INSERT INTO study_sessions (group_id, study_activities_id) VALUES (1, 1);

		INSERT INTO word_review_items (word_id, study_session_id, correct) VALUES 
		(1, 1, true),
		(2, 1, false);
	`)
	if err != nil {
		suite.T().Fatalf("Failed to insert test data: %v", err)
	}

	// Create handler with test database
	suite.h = NewHandler(suite.db)

	// Setup router
	suite.router = gin.New()
	suite.router.GET("/api/dashboard/last_study_session", suite.h.GetLastStudySession)
	suite.router.GET("/api/dashboard/study_progress", suite.h.GetStudyProgress)
	suite.router.GET("/api/dashboard/quick-stats", suite.h.GetQuickStats)
}

func (suite *DashboardTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.DB.Close()
	}
}

func TestDashboardSuite(t *testing.T) {
	suite.Run(t, new(DashboardTestSuite))
}

func (suite *DashboardTestSuite) TestGetLastStudySession() {
	// Create request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/dashboard/last_study_session", nil)
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Assert response structure
	assert.Contains(suite.T(), response, "id")
	assert.Contains(suite.T(), response, "group_id")
	assert.Contains(suite.T(), response, "created_at")
	assert.Contains(suite.T(), response, "study_activities_id")
	assert.Contains(suite.T(), response, "group_name")
}

func (suite *DashboardTestSuite) TestGetStudyProgress() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/dashboard/study_progress", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Assert response structure
	assert.Contains(suite.T(), response, "daily_stats")
	assert.Contains(suite.T(), response, "total_words_studied")
	assert.Contains(suite.T(), response, "total_available_words")
	assert.Contains(suite.T(), response, "average_accuracy")
}

func (suite *DashboardTestSuite) TestGetQuickStats() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/dashboard/quick-stats", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Assert response structure
	assert.Contains(suite.T(), response, "total_words")
	assert.Contains(suite.T(), response, "words_studied")
	assert.Contains(suite.T(), response, "total_study_sessions")
	assert.Contains(suite.T(), response, "average_accuracy")
	assert.Contains(suite.T(), response, "study_streak_days")
}
