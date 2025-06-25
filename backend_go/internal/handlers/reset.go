package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ResetHistory resets all study history
func (h *Handler) ResetHistory(c *gin.Context) {
	// Start a transaction
	tx, err := h.db.Begin()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error starting transaction")
		return
	}
	defer tx.Rollback()

	// Temporarily disable foreign key constraints
	_, err = tx.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error disabling foreign keys")
		return
	}

	// Delete all word review items
	_, err = tx.Exec("DELETE FROM word_review_items")
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error deleting word review items")
		return
	}

	// Delete all study activities
	_, err = tx.Exec("DELETE FROM study_activities")
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error deleting study activities")
		return
	}

	// Delete all study sessions
	_, err = tx.Exec("DELETE FROM study_sessions")
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error deleting study sessions")
		return
	}

	// Re-enable foreign key constraints
	_, err = tx.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error enabling foreign keys")
		return
	}

	if err := tx.Commit(); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error committing transaction")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "Study history has been reset",
		"reset_timestamp": time.Now().Format(time.RFC3339),
	})
}

// FullReset resets all data including words and groups
func (h *Handler) FullReset(c *gin.Context) {
	// Start a transaction
	tx, err := h.db.Begin()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error starting transaction")
		return
	}
	defer tx.Rollback()

	// Temporarily disable foreign key constraints
	_, err = tx.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error disabling foreign keys")
		return
	}

	// Delete all data in reverse order of dependencies
	tables := []string{
		"word_review_items",
		"study_activities",
		"study_sessions",
		"words_groups",
		"words",
		"groups",
	}

	for _, table := range tables {
		_, err = tx.Exec("DELETE FROM " + table)
		if err != nil {
			respondWithError(c, http.StatusInternalServerError, "Error deleting from "+table)
			return
		}
	}

	// Re-enable foreign key constraints
	_, err = tx.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error enabling foreign keys")
		return
	}

	if err := tx.Commit(); err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error committing transaction")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"message":         "System has been fully reset",
		"reset_timestamp": time.Now().Format(time.RFC3339),
	})
}
