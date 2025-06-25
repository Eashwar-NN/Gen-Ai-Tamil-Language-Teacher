package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lang-portal/backend_go/internal/middleware"
)

// GetStudySessions returns a paginated list of study sessions
func (h *Handler) GetStudySessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * defaultItemsPerPage

	// Get total count
	var totalItems int
	err := h.db.QueryRow("SELECT COUNT(*) FROM study_sessions").Scan(&totalItems)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error counting study sessions")
		return
	}

	// Get paginated study sessions
	rows, err := h.db.Query(`
		SELECT ss.id, sa.name as activity_name, g.name as group_name,
		       ss.created_at, COUNT(wri.id) as review_items_count
		FROM study_sessions ss
		JOIN study_activities sa ON ss.study_activities_id = sa.id
		JOIN groups g ON ss.group_id = g.id
		LEFT JOIN word_review_items wri ON ss.id = wri.study_session_id
		GROUP BY ss.id
		ORDER BY ss.created_at DESC
		LIMIT ? OFFSET ?`,
		defaultItemsPerPage, offset)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error fetching study sessions")
		return
	}
	defer rows.Close()

	var sessions []gin.H
	for rows.Next() {
		var id int
		var activityName, groupName string
		var createdAt time.Time
		var reviewItemsCount int
		if err := rows.Scan(&id, &activityName, &groupName, &createdAt, &reviewItemsCount); err != nil {
			respondWithError(c, http.StatusInternalServerError, "Error scanning study session")
			return
		}

		sessions = append(sessions, gin.H{
			"id":                 id,
			"activity_name":      activityName,
			"group_name":         groupName,
			"start_time":         createdAt,
			"review_items_count": reviewItemsCount,
		})
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Items:      sessions,
		Pagination: calculatePagination(page, defaultItemsPerPage, totalItems),
	})
}

// GetStudySession returns details of a specific study session
func (h *Handler) GetStudySession(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid study session ID")
		return
	}

	var session struct {
		ID              int       `json:"id"`
		ActivityName    string    `json:"activity_name"`
		GroupName       string    `json:"group_name"`
		StartTime       time.Time `json:"start_time"`
		ReviewItemCount int       `json:"review_items_count"`
	}

	err = h.db.QueryRow(`
		SELECT ss.id, sa.name, g.name, ss.created_at,
		       COUNT(wri.id) as review_items_count
		FROM study_sessions ss
		JOIN study_activities sa ON ss.study_activities_id = sa.id
		JOIN groups g ON ss.group_id = g.id
		LEFT JOIN word_review_items wri ON ss.id = wri.study_session_id
		WHERE ss.id = ?
		GROUP BY ss.id`,
		id).Scan(&session.ID, &session.ActivityName, &session.GroupName,
		&session.StartTime, &session.ReviewItemCount)

	if err == sql.ErrNoRows {
		respondWithError(c, http.StatusNotFound, "Study session not found")
		return
	} else if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error fetching study session")
		return
	}

	c.JSON(http.StatusOK, session)
}

// GetStudySessionWords returns all words reviewed in a specific study session
func (h *Handler) GetStudySessionWords(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid study session ID")
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize := 100 // As per specification
	offset := (page - 1) * pageSize

	// Get total count
	var totalItems int
	err = h.db.QueryRow(`
		SELECT COUNT(DISTINCT w.id)
		FROM words w
		JOIN word_review_items wri ON w.id = wri.word_id
		WHERE wri.study_session_id = ?`,
		id).Scan(&totalItems)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error counting words")
		return
	}

	// Get paginated words
	rows, err := h.db.Query(`
		SELECT w.tamil, w.romaji, w.english,
			   SUM(CASE WHEN wri.correct = 1 THEN 1 ELSE 0 END) as correct_count,
			   SUM(CASE WHEN wri.correct = 0 THEN 1 ELSE 0 END) as wrong_count
		FROM words w
		JOIN word_review_items wri ON w.id = wri.word_id
		WHERE wri.study_session_id = ?
		GROUP BY w.id
		ORDER BY w.id
		LIMIT ? OFFSET ?`,
		id, pageSize, offset)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error fetching words")
		return
	}
	defer rows.Close()

	var words []gin.H
	for rows.Next() {
		var word struct {
			Tamil        string `json:"tamil"`
			Romaji       string `json:"romaji"`
			English      string `json:"english"`
			CorrectCount int    `json:"correct_count"`
			WrongCount   int    `json:"wrong_count"`
		}

		if err := rows.Scan(&word.Tamil, &word.Romaji, &word.English, &word.CorrectCount, &word.WrongCount); err != nil {
			respondWithError(c, http.StatusInternalServerError, "Error scanning word")
			return
		}

		words = append(words, gin.H{
			"tamil":         word.Tamil,
			"romaji":        word.Romaji,
			"english":       word.English,
			"correct_count": word.CorrectCount,
			"wrong_count":   word.WrongCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items":      words,
		"pagination": calculatePagination(page, pageSize, totalItems),
	})
}

// CreateStudySession creates a new study session
func (h *Handler) CreateStudySession(c *gin.Context) {
	var req *middleware.StudySessionRequest
	if v, exists := c.Get("validated"); exists {
		req = v.(*middleware.StudySessionRequest)
	} else {
		respondWithError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// Check if group exists
	var groupExists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM groups WHERE id = ?)", req.GroupID).Scan(&groupExists)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error checking group existence")
		return
	}
	if !groupExists {
		respondWithError(c, http.StatusNotFound, "Group not found")
		return
	}

	// Check if study activity exists
	var activityExists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM study_activities WHERE id = ?)", req.StudyActivityID).Scan(&activityExists)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error checking study activity existence")
		return
	}
	if !activityExists {
		respondWithError(c, http.StatusNotFound, "Study activity not found")
		return
	}

	// Create study session
	result, err := h.db.Exec(`
		INSERT INTO study_sessions (group_id, study_activities_id, created_at)
		VALUES (?, ?, ?)`,
		req.GroupID, req.StudyActivityID, time.Now())
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error creating study session")
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error getting created study session ID")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":                id,
		"group_id":          req.GroupID,
		"study_activity_id": req.StudyActivityID,
	})
}

// ReviewWord records a word review result
func (h *Handler) ReviewWord(c *gin.Context) {
	var req *middleware.WordReviewRequest
	if v, exists := c.Get("validated"); !exists {
		respondWithError(c, http.StatusBadRequest, "Invalid request")
		return
	} else {
		req = v.(*middleware.WordReviewRequest)
	}

	sessionID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid session ID")
		return
	}

	wordID, err := strconv.Atoi(c.Param("word_id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid word ID")
		return
	}

	// Check if session exists
	var sessionExists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM study_sessions WHERE id = ?)", sessionID).Scan(&sessionExists)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error checking session existence")
		return
	}
	if !sessionExists {
		respondWithError(c, http.StatusNotFound, "Study session not found")
		return
	}

	// Check if word exists
	var wordExists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM words WHERE id = ?)", wordID).Scan(&wordExists)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error checking word existence")
		return
	}
	if !wordExists {
		respondWithError(c, http.StatusNotFound, "Word not found")
		return
	}

	// Create word review
	_, err = h.db.Exec(`
		INSERT INTO word_review_items (word_id, study_session_id, correct, created_at)
		VALUES (?, ?, ?, ?)`,
		wordID, sessionID, *req.Correct, time.Now())
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error creating word review")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"word_id":    wordID,
		"session_id": sessionID,
		"correct":    *req.Correct,
		"created_at": time.Now(),
	})
}
