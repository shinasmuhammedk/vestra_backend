package database

import (
	"fmt"
	"vestra-ecommerce/config"

	// "log"
	"sync"

	// "github.com/rs/zerolog/log"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var pgOnce sync.Once
var PgSQLDB *gorm.DB

func GetInstancepostgres( cfg *config.Config) (dba *gorm.DB) {
	pgOnce.Do(func() {
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

		// fmt.Println(":>>>>>>",dsn)
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

		//close connection - cleanup and close
		dba = db
		if err != nil {
			// log.Panic().Msgf("Error connecting to the database at %s:%s/%s", host, port, dbname)
			// log.Info().Msgf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbname)
		}

		sqlDB, err := dba.DB()
		if err != nil {
			log.Panic().Msgf("Error getting GORM DB definition")
		}
		sqlDB.SetMaxIdleConns(2)
		sqlDB.SetMaxOpenConns(10)
		//defer sqlDB.Close()
		PgSQLDB = db
		// logging.Logger.Info("Database connected successfully...")
		// log.Info().Msgf("Successfully established connection to %s:%s/%s", host, port, dbname)

	})
	PgSQLDB = dba
	return dba
}
