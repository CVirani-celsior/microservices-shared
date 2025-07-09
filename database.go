package shared

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDatabase initializes the database connection
func InitDatabase(dbPath string, models ...interface{}) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate the models
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return nil, fmt.Errorf("failed to migrate model: %w", err)
		}
	}

	logrus.WithField("database", dbPath).Info("Database initialized successfully")
	return db, nil
}
