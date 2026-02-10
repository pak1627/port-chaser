package storage

import (
	"time"

	"github.com/manson/port-chaser/internal/models"
)

// Storage defines the interface for persisting and retrieving port kill history.
// Implementations can store data in SQLite, file system, or other backends.
type Storage interface {
	// RecordKill saves a history entry when a process is killed
	RecordKill(entry models.HistoryEntry) error
	// GetHistory retrieves recent kill history, limited to the specified count
	GetHistory(limit int) ([]models.HistoryEntry, error)
	// GetKillCount returns how many times a specific port has been killed within the given days
	GetKillCount(port int, days int) (int, error)
	// GetLastKillTime returns when the specified port was last killed
	GetLastKillTime(port int) (time.Time, error)
	// Close closes the storage connection and releases resources
	Close() error
}

// Config holds configuration settings for storage implementations.
// These settings control database behavior and performance.
type Config struct {
	// DBPath is the file path to the SQLite database (empty means use default location)
	DBPath     string
	// WALEnabled enables Write-Ahead Logging for better concurrency
	WALEnabled bool
	// Timeout is the connection timeout in seconds (0 means use default)
	Timeout    int
}

// DefaultConfig returns a Config with sensible default settings.
// WAL is enabled for better performance, default timeout is 50ms.
func DefaultConfig() Config {
	return Config{
		DBPath:     "",
		WALEnabled: true,
		Timeout:    50,
	}
}
