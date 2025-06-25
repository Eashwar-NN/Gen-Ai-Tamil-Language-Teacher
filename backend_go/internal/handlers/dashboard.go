package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetLastStudySession returns information about the most recent study session
func (h *Handler) GetLastStudySession(c *gin.Context) {
	var session struct {
		ID                int       `json:"id"`
		GroupID           int       `json:"group_id"`
		CreatedAt         time.Time `json:"created_at"`
		StudyActivitiesID int       `json:"study_activities_id"`
		GroupName         string    `json:"group_name"`
	}

	err := h.db.QueryRow(`
		SELECT ss.id, ss.group_id, ss.created_at, ss.study_activities_id, g.name
		FROM study_sessions ss
		JOIN groups g ON ss.group_id = g.id
		ORDER BY ss.created_at DESC
		LIMIT 1`,
	).Scan(&session.ID, &session.GroupID, &session.CreatedAt,
		&session.StudyActivitiesID, &session.GroupName)

	if err != nil {
		// If no rows found, return empty response with 200
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{
				"id":                  0,
				"group_id":            0,
				"created_at":          time.Now(),
				"study_activities_id": 0,
				"group_name":          "",
			})
			return
		}
		respondWithError(c, http.StatusInternalServerError, "Error fetching last study session")
		return
	}

	c.JSON(http.StatusOK, session)
}

// GetStudyProgress returns study progress statistics over time
func (h *Handler) GetStudyProgress(c *gin.Context) {
	// Get daily stats for the last 30 days
	rows, err := h.db.Query(`
		WITH RECURSIVE dates(date) AS (
			SELECT date('now', '-29 days')
			UNION ALL
			SELECT date(date, '+1 day')
			FROM dates
			WHERE date < date('now')
		)
		SELECT 
			dates.date,
			COUNT(DISTINCT wri.word_id) as words_studied,
			(SELECT COUNT(*) FROM words) as total_words,
			COALESCE(AVG(CASE WHEN wri.correct THEN 100 ELSE 0 END), 0) as accuracy
		FROM dates
		LEFT JOIN study_sessions ss ON date(ss.created_at) = dates.date
		LEFT JOIN word_review_items wri ON wri.study_session_id = ss.id
		GROUP BY dates.date
		ORDER BY dates.date`,
	)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error fetching study progress")
		return
	}
	defer rows.Close()

	var dailyStats []gin.H
	var totalWordsStudied, totalAvailableWords int
	var totalAccuracy float64
	var daysWithActivity int

	for rows.Next() {
		var date string
		var wordsStudied, availableWords int
		var accuracy float64

		if err := rows.Scan(&date, &wordsStudied, &availableWords, &accuracy); err != nil {
			respondWithError(c, http.StatusInternalServerError, "Error scanning study progress")
			return
		}

		if wordsStudied > 0 {
			totalWordsStudied += wordsStudied
			totalAccuracy += accuracy
			daysWithActivity++
		}

		totalAvailableWords = availableWords // This will be the same for all rows

		dailyStats = append(dailyStats, gin.H{
			"date":                  date,
			"total_words_studied":   wordsStudied,
			"total_available_words": availableWords,
			"correct_percentage":    accuracy,
		})
	}

	averageAccuracy := 0.0
	if daysWithActivity > 0 {
		averageAccuracy = totalAccuracy / float64(daysWithActivity)
	}

	c.JSON(http.StatusOK, gin.H{
		"daily_stats":           dailyStats,
		"total_words_studied":   totalWordsStudied,
		"total_available_words": totalAvailableWords,
		"average_accuracy":      averageAccuracy,
	})
}

// GetQuickStats returns quick overview learning statistics
func (h *Handler) GetQuickStats(c *gin.Context) {
	var stats struct {
		TotalWords         int     `json:"total_words"`
		WordsStudied       int     `json:"words_studied"`
		TotalStudySessions int     `json:"total_study_sessions"`
		AverageAccuracy    float64 `json:"average_accuracy"`
		StudyStreakDays    int     `json:"study_streak_days"`
	}

	// Get total words and study sessions
	err := h.db.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM words) as total_words,
			COUNT(DISTINCT wri.word_id) as words_studied,
			COUNT(DISTINCT wri.study_session_id) as total_sessions,
			COALESCE(AVG(CASE WHEN wri.correct THEN 100 ELSE 0 END), 0) as avg_accuracy
		FROM word_review_items wri`,
	).Scan(&stats.TotalWords, &stats.WordsStudied,
		&stats.TotalStudySessions, &stats.AverageAccuracy)

	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error fetching quick stats")
		return
	}

	// Calculate study streak
	err = h.db.QueryRow(`
		WITH RECURSIVE dates(date) AS (
			SELECT date(MAX(created_at)) as date
			FROM study_sessions
			UNION ALL
			SELECT date(date, '-1 day')
			FROM dates
			WHERE EXISTS (
				SELECT 1 FROM study_sessions
				WHERE date(created_at) = date(dates.date, '-1 day')
			)
		)
		SELECT COUNT(*) FROM dates`,
	).Scan(&stats.StudyStreakDays)

	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "Error calculating study streak")
		return
	}

	c.JSON(http.StatusOK, stats)
}
