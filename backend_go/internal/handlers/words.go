package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lang-portal/backend_go/internal/middleware"
	"github.com/lang-portal/backend_go/internal/models"
)

const defaultItemsPerPage = 100

// GetWords returns a paginated list of words
func (h *Handler) GetWords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * defaultItemsPerPage

	// Get total count
	var totalItems int
	err := h.db.QueryRow("SELECT COUNT(*) FROM words").Scan(&totalItems)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error counting words")
		return
	}

	// Get paginated words
	rows, err := h.db.Query(`
		SELECT w.id, w.tamil, w.romaji, w.english,
		       COUNT(CASE WHEN wri.correct = 1 THEN 1 END) as correct_count,
		       COUNT(CASE WHEN wri.correct = 0 THEN 1 END) as wrong_count
		FROM words w
		LEFT JOIN word_review_items wri ON w.id = wri.word_id
		GROUP BY w.id
		LIMIT ? OFFSET ?`,
		defaultItemsPerPage, offset)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error fetching words")
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

// GetWord returns details of a specific word
func (h *Handler) GetWord(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid word ID")
		return
	}

	var word models.Word
	var correctCount, wrongCount int

	err = h.db.QueryRow(`
		SELECT w.id, w.tamil, w.romaji, w.english,
		       COUNT(CASE WHEN wri.correct = 1 THEN 1 END) as correct_count,
		       COUNT(CASE WHEN wri.correct = 0 THEN 1 END) as wrong_count
		FROM words w
		LEFT JOIN word_review_items wri ON w.id = wri.word_id
		WHERE w.id = ?
		GROUP BY w.id`,
		id).Scan(&word.ID, &word.Tamil, &word.Romaji, &word.English,
		&correctCount, &wrongCount)

	if err == sql.ErrNoRows {
		respondWithError(c, http.StatusNotFound, "Word not found")
		return
	} else if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error fetching word")
		return
	}

	// Get groups for this word
	rows, err := h.db.Query(`
		SELECT g.id, g.name
		FROM groups g
		JOIN words_groups wg ON g.id = wg.group_id
		WHERE wg.word_id = ?`,
		id)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error fetching word groups")
		return
	}
	defer rows.Close()

	var groups []gin.H
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.ID, &group.Name); err != nil {
			respondWithError(c, http.StatusInternalServerError, "Error scanning group")
			return
		}
		groups = append(groups, gin.H{
			"id":   group.ID,
			"name": group.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tamil":   word.Tamil,
		"romaji":  word.Romaji,
		"english": word.English,
		"stats": gin.H{
			"correct_count": correctCount,
			"wrong_count":   wrongCount,
		},
		"groups": groups,
	})
}

// CreateWord handles word creation
func (h *Handler) CreateWord(c *gin.Context) {
	var req *middleware.WordRequest
	if v, exists := c.Get("validated"); exists {
		req = v.(*middleware.WordRequest)
	} else {
		respondWithError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// Convert parts to JSON string
	parts, err := json.Marshal(req.Parts)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Invalid parts data")
		return
	}

	// Insert word into database
	result, err := h.db.Exec(`
		INSERT INTO words (tamil, romaji, english, parts)
		VALUES (?, ?, ?, ?)`,
		req.Tamil, req.Romaji, req.English, string(parts),
	)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to create word")
		return
	}

	// Get the ID of the inserted word
	id, err := result.LastInsertId()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Failed to get word ID")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      id,
		"tamil":   req.Tamil,
		"romaji":  req.Romaji,
		"english": req.English,
		"parts":   req.Parts,
	})
}
