package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lang-portal/backend_go/internal/middleware"
	"github.com/lang-portal/backend_go/internal/models"
)

// GetGroups returns a paginated list of groups
func (h *Handler) GetGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * defaultItemsPerPage

	// Get total count
	var totalItems int
	err := h.db.QueryRow("SELECT COUNT(*) FROM groups").Scan(&totalItems)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error counting groups")
		return
	}

	// Get paginated groups with word count
	rows, err := h.db.Query(`
		SELECT g.id, g.name, COUNT(wg.word_id) as word_count
		FROM groups g
		LEFT JOIN words_groups wg ON g.id = wg.group_id
		GROUP BY g.id
		LIMIT ? OFFSET ?`,
		defaultItemsPerPage, offset)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error fetching groups")
		return
	}
	defer rows.Close()

	var groups []gin.H
	for rows.Next() {
		var group models.Group
		var wordCount int
		if err := rows.Scan(&group.ID, &group.Name, &wordCount); err != nil {
			respondWithError(c, http.StatusInternalServerError, "Error scanning group")
			return
		}

		groups = append(groups, gin.H{
			"id":         group.ID,
			"name":       group.Name,
			"word_count": wordCount,
		})
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Items:      groups,
		Pagination: calculatePagination(page, defaultItemsPerPage, totalItems),
	})
}

// GetGroup returns details of a specific group
func (h *Handler) GetGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid group ID")
		return
	}

	var group models.Group
	var wordCount int

	err = h.db.QueryRow(`
		SELECT g.id, g.name, COUNT(wg.word_id) as word_count
		FROM groups g
		LEFT JOIN words_groups wg ON g.id = wg.group_id
		WHERE g.id = ?
		GROUP BY g.id`,
		id).Scan(&group.ID, &group.Name, &wordCount)

	if err == sql.ErrNoRows {
		respondWithError(c, http.StatusNotFound, "Group not found")
		return
	} else if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error fetching group")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":   group.ID,
		"name": group.Name,
		"stats": gin.H{
			"total_word_count": wordCount,
		},
	})
}

// GetGroupWords returns all words in a specific group
func (h *Handler) GetGroupWords(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid group ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * defaultItemsPerPage

	// Get total count
	var totalItems int
	err = h.db.QueryRow(`
		SELECT COUNT(*)
		FROM words w
		JOIN words_groups wg ON w.id = wg.word_id
		WHERE wg.group_id = ?`,
		id).Scan(&totalItems)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error counting group words")
		return
	}

	// Get paginated words
	rows, err := h.db.Query(`
		SELECT w.id, w.tamil, w.romaji, w.english,
		       COUNT(CASE WHEN wri.correct = 1 THEN 1 END) as correct_count,
		       COUNT(CASE WHEN wri.correct = 0 THEN 1 END) as wrong_count
		FROM words w
		JOIN words_groups wg ON w.id = wg.word_id
		LEFT JOIN word_review_items wri ON w.id = wri.word_id
		WHERE wg.group_id = ?
		GROUP BY w.id
		LIMIT ? OFFSET ?`,
		id, defaultItemsPerPage, offset)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error fetching group words")
		return
	}
	defer rows.Close()

	var words []gin.H
	for rows.Next() {
		var word models.Word
		var correctCount, wrongCount int
		if err := rows.Scan(&word.ID, &word.Tamil, &word.Romaji, &word.English,
			&correctCount, &wrongCount); err != nil {
			respondWithError(c, http.StatusInternalServerError, "Error scanning word")
			return
		}

		words = append(words, gin.H{
			"tamil":         word.Tamil,
			"romaji":        word.Romaji,
			"english":       word.English,
			"correct_count": correctCount,
			"wrong_count":   wrongCount,
		})
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Items:      words,
		Pagination: calculatePagination(page, defaultItemsPerPage, totalItems),
	})
}

// GetGroupStudySessions returns all study sessions for a specific group
func (h *Handler) GetGroupStudySessions(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid group ID")
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
		SELECT COUNT(*)
		FROM study_sessions ss
		WHERE ss.group_id = ?`,
		id).Scan(&totalItems)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error counting study sessions")
		return
	}

	// Get paginated study sessions
	rows, err := h.db.Query(`
		SELECT ss.id, sa.name as activity_name, g.name as group_name,
			   ss.created_at as start_time,
			   (SELECT created_at 
				FROM word_review_items 
				WHERE study_session_id = ss.id 
				ORDER BY created_at DESC 
				LIMIT 1) as end_time,
			   COUNT(wri.id) as review_items_count
		FROM study_sessions ss
		JOIN study_activities sa ON ss.study_activities_id = sa.id
		JOIN groups g ON ss.group_id = g.id
		LEFT JOIN word_review_items wri ON ss.id = wri.study_session_id
		WHERE ss.group_id = ?
		GROUP BY ss.id
		ORDER BY ss.created_at DESC
		LIMIT ? OFFSET ?`,
		id, pageSize, offset)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error fetching study sessions")
		return
	}
	defer rows.Close()

	var sessions []gin.H
	for rows.Next() {
		var session struct {
			ID              int     `json:"id"`
			ActivityName    string  `json:"activity_name"`
			GroupName       string  `json:"group_name"`
			StartTime       string  `json:"start_time"`
			EndTime         *string `json:"end_time"`
			ReviewItemCount int     `json:"review_items_count"`
		}

		if err := rows.Scan(
			&session.ID, &session.ActivityName, &session.GroupName,
			&session.StartTime, &session.EndTime, &session.ReviewItemCount,
		); err != nil {
			respondWithError(c, http.StatusInternalServerError, "Error scanning study session")
			return
		}

		sessions = append(sessions, gin.H{
			"id":                 session.ID,
			"activity_name":      session.ActivityName,
			"group_name":         session.GroupName,
			"start_time":         session.StartTime,
			"end_time":           session.EndTime,
			"review_items_count": session.ReviewItemCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items":      sessions,
		"pagination": calculatePagination(page, pageSize, totalItems),
	})
}

// CreateGroup handles group creation
func (h *Handler) CreateGroup(c *gin.Context) {
	var req *middleware.GroupRequest
	if v, exists := c.Get("validated"); exists {
		req = v.(*middleware.GroupRequest)
	} else {
		respondWithError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// Insert group into database
	result, err := h.db.Exec(`
		INSERT INTO groups (name)
		VALUES (?)`,
		req.Name,
	)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to create group")
		return
	}

	// Get the ID of the inserted group
	id, err := result.LastInsertId()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to get group ID")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":   id,
		"name": req.Name,
	})
}
