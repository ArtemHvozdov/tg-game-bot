package storage_db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// InitDB initializate database SQLite with path dbPath
func InitDB(dbPath string) (*sql.DB, error) {
	// Connect to database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Error connection database: %v", err)
		return nil, err
	}

	// Check connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Error checking connect to database: %v", err)
		return nil, err
	}

	log.Println("The database has been initialized successfully.")
	return db, nil
}

// CloseDB close connect to database
func CloseDB(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("The database connection was closed successfully.")
		}
	}
}