package database

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/jmoiron/sqlx"
)

func Migrate(db DatabaseProvider, migrationGlobs []string) error {
	migrations := resolveMigrations(migrationGlobs)
	if len(migrations) == 0 {
		return nil
	}

	return runMigrations(db.Get(), migrations)
}

type migration struct {
	version string
	path    string
}

func resolveMigrations(globs []string) []migration {
	result := make([]migration, 0, len(globs))

	for _, glob := range globs {
		files, err := filepath.Glob(glob)
		if err != nil {
			panic(err)
		}

		for _, file := range files {
			log.Println("Migration file", file)
			base := filepath.Base(file)

			result = append(result, migration{
				version: base,
				path:    file,
			})
		}
	}

	return result
}

func setupMigrations(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id		INTEGER PRIMARY KEY,
			version TEXT
		);
	`)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func getCurrentMigration(db *sqlx.DB) (string, error) {
	var version string
	row := db.QueryRowx("SELECT version FROM migrations WHERE id = 1")
	err := row.Scan(&version)
	if err == sql.ErrNoRows {
		_, err = db.Exec(`INSERT INTO migrations (id, version) VALUES (1, "NA")`)
		if err != nil {
			return "", err
		}

		return "NA", nil
	}

	if err != nil {
		return "", err
	}

	return version, nil
}

// TODO: Binary Search
func findLastMigrationIndex(migrations []migration, version string) (int, error) {
	if version == "NA" {
		return -1, nil
	}

	for i, m := range migrations {
		if m.version == version {
			return i, nil
		}
	}

	return -1, fmt.Errorf("Index not found for version %s", version)
}

func runMigrations(db *sqlx.DB, migrations []migration) error {
	var err error

	// ensure migrations table exists
	err = setupMigrations(db)
	if err != nil {
		return err
	}

	version, err := getCurrentMigration(db)
	if err != nil {
		return err
	}

	sort.SliceStable(migrations, func(a, b int) bool {
		return migrations[a].version < migrations[b].version
	})

	// Get current version
	i, err := findLastMigrationIndex(migrations, version)
	if err != nil {
		return err
	}

	i++
	for i < len(migrations) {
		var queryFile *os.File
		var query []byte
		var tx *sql.Tx

		migration := migrations[i]
		queryFile, err = os.Open(migration.path)
		if err != nil {
			break
		}

		query, err = io.ReadAll(queryFile)
		if err != nil {
			break
		}

		tx, err = db.Begin()
		if err != nil {
			break
		}

		_, err = tx.Exec(string(query))

		if err != nil {
			tx.Rollback()
			break
		}

		err = tx.Commit()
		if err != nil {
			break
		}

		i++
	}

	lastSuccessful := i - 1
	if lastSuccessful >= 0 {
		_, err = db.Exec(
			`UPDATE migrations SET version = ? WHERE id = 1`,
			migrations[lastSuccessful].version,
		)
	}

	if err != nil {
		return err
	}

	return nil
}
