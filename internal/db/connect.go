package db

import (
	"auth/internal/config"
	"fmt"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type AuthDatabase struct {
	Database *gorm.DB
}

func ConnectDB(envConf *config.Config) *AuthDatabase {
	connectionString := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		envConf.DB.Host,
		envConf.DB.User,
		envConf.DB.Password,
		envConf.DB.Name,
		envConf.DB.Port,
		envConf.DB.SSLMode,
	)

	logLevel := logger.Info
	if envConf.ProductionType == "prod" {
		logLevel = logger.Error
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	database, err := gorm.Open(postgres.Open(connectionString), gormConfig)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	sqlDB, err := database.DB()
	if err != nil {
		panic(fmt.Errorf("failed to get sql.DB from gorm DB: %w", err))
	}

	maxIdleConns := envConf.DB.MaxIdleConns
	maxOpenConns := envConf.DB.MaxOpenConns
	connMaxLifetime := envConf.DB.ConnMaxLifetime

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)

	if err := sqlDB.Ping(); err != nil {
		panic(fmt.Errorf("failed to ping database: %w", err))
	}

	log.Info().Msg("connected to the database successfully")

	return &AuthDatabase{
		Database: database,
	}
}
