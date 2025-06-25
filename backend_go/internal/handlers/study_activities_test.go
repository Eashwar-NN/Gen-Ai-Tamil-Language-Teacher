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

type StudyActivitiesTestSuite struct {
	suite.Suite
	router *gin.Engine
	h      *Handler
	db     *models.DB
}

func (suite *StudyActivitiesTestSuite) SetupTest() {
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

		-- Then, create study sessions
		INSERT INTO study_sessions (id, group_id) VALUES (1, 1);

		-- Then, create study activities
		INSERT INTO study_activities (id, study_session_id, group_id) VALUES 
		(1, 1, 1),
		(2, 1, 2);

		-- Update study session with study activity
		UPDATE study_sessions SET study_activities_id = 1 WHERE id = 1;
	`)
	if err != nil {
		suite.T().Fatalf("Failed to insert test data: %v", err)
	}

	// Create handler with test database
	suite.h = NewHandler(suite.db)

	// Setup router with proper parameter handling
	suite.router = gin.New()
	api := suite.router.Group("/api")
	api.GET("/study_activities/:id", suite.h.GetStudyActivity)
	api.GET("/study_activities/:id/study_sessions", suite.h.GetStudyActivitySessions)
	api.POST("/study_activities", suite.h.CreateStudyActivity)
}

func (suite *StudyActivitiesTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.DB.Close()
	}
}

func TestStudyActivitiesTestSuite(t *testing.T) {
	suite.Run(t, new(StudyActivitiesTestSuite))
}

func (suite *StudyActivitiesTestSuite) TestGetStudyActivity() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/study_activities/1", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		ThumbnailURL string `json:"thumbnail_url"`
		Description  string `json:"description"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), 1, response.ID)
	assert.NotEmpty(suite.T(), response.Name)
	assert.NotEmpty(suite.T(), response.ThumbnailURL)
	assert.NotEmpty(suite.T(), response.Description)
}

func (suite *StudyActivitiesTestSuite) TestGetStudyActivitySessions() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/study_activities/1/study_sessions", nil)
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
}

func (suite *StudyActivitiesTestSuite) TestCreateStudyActivity() {
	reqBody := map[string]interface{}{
		"group_id": 1,
		"name":     "Vocabulary Quiz",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/study_activities", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response struct {
		ID      int `json:"id"`
		GroupID int `json:"group_id"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.NotZero(suite.T(), response.ID)
	assert.Equal(suite.T(), 1, response.GroupID)
}
