package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lang-portal/backend_go/internal/middleware"
)

// GetStudyActivity returns details of a specific study activity
func (h *Handler) GetStudyActivity(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid study activity ID")
		return
	}

	var activity struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		ThumbnailURL string `json:"thumbnail_url"`
		Description  string `json:"description"`
	}

	err = h.db.QueryRow(`
		SELECT id, name, thumbnail_url, description
		FROM study_activities
		WHERE id = ?`,
		id).Scan(&activity.ID, &activity.Name, &activity.ThumbnailURL, &activity.Description)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(c, http.StatusNotFound, "Study activity not found")
			return
		}
		respondWithError(c, http.StatusInternalServerError, "Error fetching study activity")
		return
	}

	c.JSON(http.StatusOK, activity)
}

// GetStudyActivitySessions returns all study sessions for a specific activity
func (h *Handler) GetStudyActivitySessions(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid study activity ID")
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
		WHERE ss.study_activities_id = ?`,
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
		WHERE ss.study_activities_id = ?
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

// CreateStudyActivity creates a new study activity
func (h *Handler) CreateStudyActivity(c *gin.Context) {
	var req *middleware.StudyActivityRequest
	if v, exists := c.Get("validated"); exists {
		req = v.(*middleware.StudyActivityRequest)
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

	// Create study activity
	result, err := h.db.Exec(`
		INSERT INTO study_activities (group_id, name, thumbnail_url, description, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		req.GroupID, req.Name, "https://example.com/thumbnail.jpg", "Practice your vocabulary with flashcards")
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error creating study activity")
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error getting created study activity ID")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       id,
		"group_id": req.GroupID,
	})
}
