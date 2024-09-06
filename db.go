package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// DBConfig struct holds the database connection configuration
type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
}

// connectToDatabase attempts to connect to the PostgreSQL database with the given configuration
func connectToDatabase(config DBConfig) (*sqlx.DB, error) {
	// Connection string for PostgreSQL including the port
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	// Try connecting to the database
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %v", err)
	}

	// Test connection (ping the database)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping the database: %v", err)
	}

	return db, nil
}

// fetchTables fetches the list of tables in the connected database
func fetchTables(db *sqlx.DB) ([]string, error) {
	query := `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'`
	var tables []string

	err := db.Select(&tables, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tables: %v", err)
	}

	return tables, nil
}

// fetchRowCount fetches the row count for a given table
func fetchRowCount(db *sqlx.DB, tableName string) (int, error) {
	var rowCount int
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, tableName)

	err := db.Get(&rowCount, query)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch row count for table %s: %v", tableName, err)
	}

	return rowCount, nil
}

