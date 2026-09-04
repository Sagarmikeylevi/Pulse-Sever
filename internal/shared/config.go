package shared

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	App AppConfig
	DB  DBConfig
	JWT JWTConfig
}

type AppConfig struct {
	Port string `env:"APP_PORT" envDefault:"8080"`
	Env  string `env:"APP_ENV" envDefault:"development"`
}

type JWTConfig struct {
	Secret             string `env:"JWT_SECRET,required"`
	AccessTokenExpiry  int    `env:"JWT_ACCESS_EXPIRY" envDefault:"15"`
	RefreshTokenExpiry int    `env:"JWT_REFRESH_EXPIRY" envDefault:"43200"`
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

	appConfig := AppConfig{}
	if err := env.Parse(&appConfig); err != nil {
		return nil, err
	}

	dbConfig := DBConfig{}
	if err := env.Parse(&dbConfig); err != nil {
		return nil, err
	}

	jwtConfig := JWTConfig{}
	if err := env.Parse(&jwtConfig); err != nil {
		return nil, err
	}

	return &Config{
		App: appConfig,
		DB:  dbConfig,
		JWT: jwtConfig,
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
