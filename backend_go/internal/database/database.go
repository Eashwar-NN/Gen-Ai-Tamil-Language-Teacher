package database

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lang-portal/backend_go/internal/models"
)

const (
	DBPath = "words.db"
)

// SeedData represents the structure of our seed JSON files
type SeedData struct {
	GroupName string `json:"group_name"`
	Words     []struct {
		Tamil   string          `json:"tamil"`
		Romaji  string          `json:"romaji"`
		English string          `json:"english"`
		Parts   json.RawMessage `json:"parts"`
	} `json:"words"`
}

// InitDB initializes the database connection
func InitDB() (*models.DB, error) {
	db, err := models.NewDB(DBPath)
	if err != nil {
		return nil, fmt.Errorf("error initializing database: %v", err)
	}
	return db, nil
}

// RunMigrations runs all migration files in the migrations directory
func RunMigrations(db *models.DB) error {
	files, err := ioutil.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("error reading migrations directory: %v", err)
	}

	var migrationFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	sort.Strings(migrationFiles)

	for _, file := range migrationFiles {
		content, err := ioutil.ReadFile(filepath.Join("migrations", file))
		if err != nil {
			return fmt.Errorf("error reading migration file %s: %v", file, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("error executing migration file %s: %v", file, err)
		}
	}

	return nil
}

// SeedDatabase seeds the database with initial data
func SeedDatabase(db *models.DB) error {
	files, err := ioutil.ReadDir("seeds")
	if err != nil {
		return fmt.Errorf("error reading seeds directory: %v", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			if err := processSeedFile(db, filepath.Join("seeds", file.Name())); err != nil {
				return fmt.Errorf("error processing seed file %s: %v", file.Name(), err)
			}
		}
	}

	return nil
}

// processSeedFile processes a single seed file
func processSeedFile(db *models.DB, filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	var seedData SeedData
	if err := json.Unmarshal(content, &seedData); err != nil {
		return fmt.Errorf("error parsing seed file: %v", err)
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert or get group
	var groupID int64
	err = tx.QueryRow(
		"INSERT OR IGNORE INTO groups (name) VALUES (?) RETURNING id",
		seedData.GroupName,
	).Scan(&groupID)
	if err != nil {
		// If the group already exists, get its ID
		err = tx.QueryRow(
			"SELECT id FROM groups WHERE name = ?",
			seedData.GroupName,
		).Scan(&groupID)
		if err != nil {
			return fmt.Errorf("error getting group ID: %v", err)
		}
	}

	// Insert words and create word-group associations
	for _, word := range seedData.Words {
		var wordID int64
		err = tx.QueryRow(
			`INSERT INTO words (tamil, romaji, english, parts)
			VALUES (?, ?, ?, ?)
			RETURNING id`,
			word.Tamil, word.Romaji, word.English, word.Parts,
		).Scan(&wordID)
		if err != nil {
			return fmt.Errorf("error inserting word: %v", err)
		}

		// Create word-group association
		_, err = tx.Exec(
			"INSERT INTO words_groups (word_id, group_id) VALUES (?, ?)",
			wordID, groupID,
		)
		if err != nil {
			return fmt.Errorf("error creating word-group association: %v", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}
