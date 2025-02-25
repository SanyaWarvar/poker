package server

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type PostgresConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

func (cfg *PostgresConfig) generateDataSourceName() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.DBName,
		cfg.Password,
		cfg.SSLMode,
	)
}

func NewPostgresDB(cfg PostgresConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", cfg.generateDataSourceName())
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil

}
