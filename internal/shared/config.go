package shared

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	DB DBConfig
}

type DBConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER" envDefault:"pulse"`
	Password string `env:"DB_PASSWORD,required"`
	Name     string `env:"DB_NAME" envDefault:"pulse_dev"`
	SSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	dbConfig := DBConfig{}

	if err := env.Parse(&dbConfig); err != nil {
		return nil, err
	}

	return &Config{
		DB: dbConfig,
	}, nil
}

func (db DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host,
		db.Port,
		db.User,
		db.Password,
		db.Name,
		db.SSLMode,
	)
}
