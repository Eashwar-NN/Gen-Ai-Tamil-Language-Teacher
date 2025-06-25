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

type GroupsTestSuite struct {
	suite.Suite
	router *gin.Engine
	h      *Handler
	db     *models.DB
}

func (suite *GroupsTestSuite) SetupTest() {
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
	`)
	if err != nil {
		suite.T().Fatalf("Failed to insert test data: %v", err)
	}

	// Create handler with test database
	suite.h = NewHandler(suite.db)

	// Setup router with proper parameter handling
	suite.router = gin.New()
	api := suite.router.Group("/api")
	api.GET("/groups", suite.h.GetGroups)
	api.GET("/groups/:id", suite.h.GetGroup)
	api.GET("/groups/:id/words", suite.h.GetGroupWords)
	api.GET("/groups/:id/study_sessions", suite.h.GetGroupStudySessions)
}

func (suite *GroupsTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.DB.Close()
	}
}

func TestGroupsTestSuite(t *testing.T) {
	suite.Run(t, new(GroupsTestSuite))
}

func (suite *GroupsTestSuite) TestGetGroups() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/groups", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response struct {
		Items []struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			WordCount int    `json:"word_count"`
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
	assert.Equal(suite.T(), "Basic Greetings", response.Items[0].Name)
	assert.Equal(suite.T(), 2, response.Items[0].WordCount)
}

func (suite *GroupsTestSuite) TestGetGroup() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/groups/1", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Stats struct {
			TotalWordCount int `json:"total_word_count"`
		} `json:"stats"`
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), 1, response.ID)
	assert.Equal(suite.T(), "Basic Greetings", response.Name)
	assert.Equal(suite.T(), 2, response.Stats.TotalWordCount)
}

func (suite *GroupsTestSuite) TestGetGroupWords() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/groups/1/words", nil)
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
	assert.Equal(suite.T(), "வணக்கம்", response.Items[0].Tamil)
}

func (suite *GroupsTestSuite) TestGetGroupStudySessions() {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/groups/1/study_sessions", nil)
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
	assert.Equal(suite.T(), "Basic Greetings", response.Items[0].GroupName)
}
