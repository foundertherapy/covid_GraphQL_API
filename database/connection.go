package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

func ConnectDB() (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", "covid.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	// create tables all at once in a transaction:
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	err = CreateTables(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CreateTables(tx *sql.Tx) error {
	err := createCountryTable(tx)
	if err != nil {
		return err
	}

	err = createCovidStatisticTable(tx)
	if err != nil {
		return err
	}

	err = createUserTable(tx)
	if err != nil {
		return err
	}

	err = createUserMonitoredCountriesTable(tx)
	if err != nil {
		return err
	}

	return nil
}

func createCountryTable(tx *sql.Tx) error {
	createCountryTable := `
		CREATE TABLE IF NOT EXISTS countries (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		code TEXT NOT NULL UNIQUE
	);`
	_, err := tx.Exec(createCountryTable)
	return err
}

func createCovidStatisticTable(tx *sql.Tx) error {
	createCovidStatisticTable := `
		CREATE TABLE IF NOT EXISTS covid_statistics (
		id INTEGER PRIMARY KEY,
		country_id INTEGER NOT NULL,
		date TEXT NOT NULL,
		confirmed INTEGER NOT NULL,
		recovered INTEGER NOT NULL,
		deaths INTEGER NOT NULL,
		FOREIGN KEY (country_id) REFERENCES countries (id) ON DELETE CASCADE
	);`
	_, err := tx.Exec(createCovidStatisticTable)
	return err
}

func createUserTable(tx *sql.Tx) error {
	createUserTable := `
		CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		salt BLOB NOT NULL
	);`
	_, err := tx.Exec(createUserTable)
	return err
}

func createUserMonitoredCountriesTable(tx *sql.Tx) error {
	createUserMonitoredCountriesTable := `
		CREATE TABLE IF NOT EXISTS user_monitored_countries (
		user_id INTEGER NOT NULL,
		country_id INTEGER NOT NULL,
		PRIMARY KEY (user_id, country_id),
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
		FOREIGN KEY (country_id) REFERENCES countries (id) ON DELETE CASCADE
	);`
	_, err := tx.Exec(createUserMonitoredCountriesTable)
	return err
}

func NewDB(db *sql.DB) *DB {
	return &DB{db: db}
}
