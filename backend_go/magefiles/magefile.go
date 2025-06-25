// +build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/lang-portal/backend_go/internal/database"
)

// DB namespace for database operations
type DB mg.Namespace

// Init initializes the database
func (DB) Init() error {
	db, err := database.InitDB()
	if err != nil {
		return err
	}
	defer db.Close()
	return nil
}

// Migrate runs database migrations
func (DB) Migrate() error {
	db, err := database.InitDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return database.RunMigrations(db)
}

// Seed seeds the database with initial data
func (DB) Seed() error {
	db, err := database.InitDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return database.SeedDatabase(db)
}

// Reset resets the database by running migrations and seeds
func (DB) Reset() error {
	db, err := database.InitDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		return err
	}

	return database.SeedDatabase(db)
} 