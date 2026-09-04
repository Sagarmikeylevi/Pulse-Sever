package shared

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func newMigrateInstance(dbConfig DBConfig) (*migrate.Migrate, error) {
	dbURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
		dbConfig.SSLMode,
	)

	m, err := migrate.New("file://migration", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m, nil
}

func MigrateUp(dbConfig DBConfig) error {
	m, err := newMigrateInstance(dbConfig)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("migrations: no new migrations to apply")
		return nil
	}
	if err != nil {
		return fmt.Errorf("migration up failed: %w", err)
	}

	log.Println("migrations: all migrations applied successfully")
	return nil
}

func MigrateDown(dbConfig DBConfig) error {
	m, err := newMigrateInstance(dbConfig)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Steps(-1)
	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("migrations: no migrations to rollback")
		return nil
	}
	if err != nil {
		return fmt.Errorf("migration down failed: %w", err)
	}

	log.Println("migrations: rolled back one migration")
	return nil
}

func MigrateDownAll(dbConfig DBConfig) error {
	m, err := newMigrateInstance(dbConfig)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Down()
	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("migrations: no migrations to rollback")
		return nil
	}
	if err != nil {
		return fmt.Errorf("migration down-all failed: %w", err)
	}

	log.Println("migrations: all migrations rolled back")
	return nil
}

func MigrateCreate(name string) error {
	if name == "" {
		return fmt.Errorf("migration name is required")
	}

	// Find next sequence number by counting existing migration pairs
	entries, err := os.ReadDir("migration")
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	seq := 1
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".up.sql") {
			seq++
		}
	}

	upFile := fmt.Sprintf("migration/%06d_%s.up.sql", seq, name)
	downFile := fmt.Sprintf("migration/%06d_%s.down.sql", seq, name)

	if err := os.WriteFile(upFile, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to create up migration: %w", err)
	}
	if err := os.WriteFile(downFile, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to create down migration: %w", err)
	}

	log.Printf("migrations: created %s", upFile)
	log.Printf("migrations: created %s", downFile)
	return nil
}
