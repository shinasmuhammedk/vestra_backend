package database

import (
	"fmt"
	"sync"

	"vestra-ecommerce/config"
	"vestra-ecommerce/utils/logging"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var pgOnce sync.Once          // Ensures database initializes only once
var PgSQLDB *gorm.DB          // Holds the PostgreSQL database instance

// GetInstancepostgres initializes and returns PostgreSQL DB singleton
func GetInstancepostgres(cfg *config.Config) (dba *gorm.DB) {

	pgOnce.Do(func() {

		// Build PostgreSQL DSN string
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
			cfg.DB.Host,
			cfg.DB.User,
			cfg.DB.Password,
			cfg.DB.Name,
			cfg.DB.Port,
			cfg.DB.SSLMode,
			cfg.DB.TimeZone,
		)

		logging.Debug.Println("Connecting to PostgreSQL database")

		// Open GORM PostgreSQL connection
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			logging.Error.Println("Failed to connect to database:", err)
			return
		}

		// Get underlying sql.DB for connection pool configuration
		sqlDB, err := db.DB()
		if err != nil {
			logging.Error.Println("Failed to get sql.DB instance:", err)
			return
		}

		// Configure database connection pool
		sqlDB.SetMaxIdleConns(2)
		sqlDB.SetMaxOpenConns(10)

		PgSQLDB = db
		dba = db

		logging.Debug.Println("PostgreSQL database connected successfully")
	})

	return PgSQLDB
}
