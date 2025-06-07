// Package db provides database driver abstraction and connection management
package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jrschumacher/dis.quest/internal/config"
	"github.com/jrschumacher/dis.quest/internal/logger"
	
	// Database drivers
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/lib/pq"
)

// DatabaseDriver represents the type of database driver
type DatabaseDriver string

// Database driver constants
const (
	SQLite     DatabaseDriver = "sqlite3"
	PostgreSQL DatabaseDriver = "postgres"
)

// DatabaseConfig holds database-specific configuration
type DatabaseConfig struct {
	Driver          DatabaseDriver
	ConnectionString string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// DetectDriver determines the database driver from the connection string
func DetectDriver(connectionString string) DatabaseDriver {
	connectionString = strings.ToLower(connectionString)
	
	switch {
	case strings.HasPrefix(connectionString, "postgres://") || 
		 strings.HasPrefix(connectionString, "postgresql://") ||
		 strings.Contains(connectionString, "host="):
		return PostgreSQL
	case strings.HasSuffix(connectionString, ".db") ||
		 strings.Contains(connectionString, "file:") ||
		 connectionString == ":memory:" ||
		 !strings.Contains(connectionString, "://"):
		return SQLite
	default:
		// Default to SQLite for simple paths
		return SQLite
	}
}

// OpenDatabase opens a database connection with the appropriate driver and settings
func OpenDatabase(cfg *config.Config) (*sql.DB, DatabaseDriver, error) {
	dbConfig := DatabaseConfig{
		Driver:           DetectDriver(cfg.DatabaseURL),
		ConnectionString: cfg.DatabaseURL,
		MaxOpenConns:     25,  // Default for production
		MaxIdleConns:     5,   // Default for production
		ConnMaxLifetime:  5 * time.Minute,
	}
	
	// Adjust settings based on driver and environment
	switch dbConfig.Driver {
	case SQLite:
		// SQLite doesn't benefit from connection pooling
		dbConfig.MaxOpenConns = 1
		dbConfig.MaxIdleConns = 1
		dbConfig.ConnMaxLifetime = 0 // No limit for SQLite
		
		// Add SQLite-specific parameters if not present
		if !strings.Contains(dbConfig.ConnectionString, "?") {
			dbConfig.ConnectionString += "?_busy_timeout=10000&_journal_mode=WAL&_foreign_keys=on"
		}
		
	case PostgreSQL:
		// Use connection pooling for PostgreSQL
		if cfg.AppEnv == "development" {
			dbConfig.MaxOpenConns = 10
			dbConfig.MaxIdleConns = 2
		}
	}
	
	logger.Info("Opening database connection", 
		"driver", string(dbConfig.Driver),
		"maxOpenConns", dbConfig.MaxOpenConns,
		"maxIdleConns", dbConfig.MaxIdleConns)
	
	// Open the database connection
	db, err := sql.Open(string(dbConfig.Driver), dbConfig.ConnectionString)
	if err != nil {
		return nil, dbConfig.Driver, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Configure connection pool
	db.SetMaxOpenConns(dbConfig.MaxOpenConns)
	db.SetMaxIdleConns(dbConfig.MaxIdleConns)
	db.SetConnMaxLifetime(dbConfig.ConnMaxLifetime)
	
	// Test the connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, dbConfig.Driver, fmt.Errorf("failed to ping database: %w", err)
	}
	
	// Apply driver-specific initialization
	if err := initializeDatabase(db, dbConfig.Driver); err != nil {
		_ = db.Close()
		return nil, dbConfig.Driver, fmt.Errorf("failed to initialize database: %w", err)
	}
	
	return db, dbConfig.Driver, nil
}

// initializeDatabase applies driver-specific initialization
func initializeDatabase(db *sql.DB, driver DatabaseDriver) error {
	switch driver {
	case SQLite:
		// Enable foreign key constraints for SQLite
		if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
			return fmt.Errorf("failed to enable foreign keys: %w", err)
		}
		
		// Set other useful SQLite pragmas
		pragmas := []string{
			"PRAGMA journal_mode = WAL",
			"PRAGMA synchronous = NORMAL", 
			"PRAGMA cache_size = -64000", // 64MB cache
			"PRAGMA temp_store = MEMORY",
		}
		
		for _, pragma := range pragmas {
			if _, err := db.Exec(pragma); err != nil {
				logger.Warn("Failed to set SQLite pragma", "pragma", pragma, "error", err)
			}
		}
		
	case PostgreSQL:
		// PostgreSQL-specific initialization if needed
		// For example, setting timezone or other session variables
		if _, err := db.Exec("SET timezone = 'UTC'"); err != nil {
			logger.Warn("Failed to set PostgreSQL timezone", "error", err)
		}
	}
	
	return nil
}

// GetPlaceholder returns the appropriate SQL placeholder for the driver
func GetPlaceholder(driver DatabaseDriver, position int) string {
	switch driver {
	case PostgreSQL:
		return fmt.Sprintf("$%d", position)
	case SQLite:
		return "?"
	default:
		return "?"
	}
}

// GetLimitOffset returns the appropriate LIMIT/OFFSET syntax for the driver
func GetLimitOffset(driver DatabaseDriver, limit, offset int64) string {
	switch driver {
	case PostgreSQL, SQLite:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	default:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}
}

// SupportsReturning returns whether the driver supports RETURNING clauses
func SupportsReturning(driver DatabaseDriver) bool {
	switch driver {
	case PostgreSQL:
		return true
	case SQLite:
		return true // SQLite 3.35+ supports RETURNING
	default:
		return false
	}
}