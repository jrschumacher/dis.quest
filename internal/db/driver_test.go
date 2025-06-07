package db

import (
	"testing"
)

func TestDetectDriver(t *testing.T) {
	tests := []struct {
		name           string
		connectionString string
		expectedDriver DatabaseDriver
	}{
		{
			name:           "SQLite file path",
			connectionString: "./test.db",
			expectedDriver:   SQLite,
		},
		{
			name:           "SQLite file URL",
			connectionString: "file:test.db",
			expectedDriver:   SQLite,
		},
		{
			name:           "SQLite memory",
			connectionString: ":memory:",
			expectedDriver:   SQLite,
		},
		{
			name:           "PostgreSQL URL",
			connectionString: "postgres://user:pass@localhost:5432/dbname",
			expectedDriver:   PostgreSQL,
		},
		{
			name:           "PostgreSQL alternative URL",
			connectionString: "postgresql://user:pass@localhost:5432/dbname",
			expectedDriver:   PostgreSQL,
		},
		{
			name:           "PostgreSQL with host parameter",
			connectionString: "host=localhost user=test dbname=test",
			expectedDriver:   PostgreSQL,
		},
		{
			name:           "Simple path defaults to SQLite",
			connectionString: "mydb",
			expectedDriver:   SQLite,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := DetectDriver(tt.connectionString)
			if driver != tt.expectedDriver {
				t.Errorf("DetectDriver(%q) = %v, want %v", 
					tt.connectionString, driver, tt.expectedDriver)
			}
		})
	}
}

func TestGetPlaceholder(t *testing.T) {
	tests := []struct {
		name     string
		driver   DatabaseDriver
		position int
		expected string
	}{
		{
			name:     "PostgreSQL placeholder",
			driver:   PostgreSQL,
			position: 1,
			expected: "$1",
		},
		{
			name:     "PostgreSQL placeholder position 3",
			driver:   PostgreSQL,
			position: 3,
			expected: "$3",
		},
		{
			name:     "SQLite placeholder",
			driver:   SQLite,
			position: 1,
			expected: "?",
		},
		{
			name:     "SQLite placeholder any position",
			driver:   SQLite,
			position: 5,
			expected: "?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPlaceholder(tt.driver, tt.position)
			if result != tt.expected {
				t.Errorf("GetPlaceholder(%v, %d) = %q, want %q",
					tt.driver, tt.position, result, tt.expected)
			}
		})
	}
}

func TestSupportsReturning(t *testing.T) {
	tests := []struct {
		name     string
		driver   DatabaseDriver
		expected bool
	}{
		{
			name:     "PostgreSQL supports RETURNING",
			driver:   PostgreSQL,
			expected: true,
		},
		{
			name:     "SQLite supports RETURNING",
			driver:   SQLite,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SupportsReturning(tt.driver)
			if result != tt.expected {
				t.Errorf("SupportsReturning(%v) = %v, want %v",
					tt.driver, result, tt.expected)
			}
		})
	}
}