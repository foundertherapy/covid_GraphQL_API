package database

import (
	"database/sql"
	"fmt"
)

func (d *DB) CreateCovidStatistic(countryID int, date string, confirmed int, recovered int, deaths int) (CovidStatistic, error) {
	result, err := d.db.Exec("INSERT INTO covid_statistics (country_id, date, confirmed, recovered, deaths) VALUES (?, ?, ?, ?, ?)", countryID, date, confirmed, recovered, deaths)
	if err != nil {
		return CovidStatistic{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return CovidStatistic{}, err
	}

	return CovidStatistic{
		ID:        int(id),
		CountryID: countryID,
		Date:      date,
		Confirmed: confirmed,
		Recovered: recovered,
		Deaths:    deaths,
	}, nil
}

func (d *DB) CreateCountry(name string, code string) (Country, bool, error) {
	ifExists := checkIfCountryExists(d.db, name)
	if ifExists {
		return Country{}, true, nil
	}

	insertCountryQuery := "INSERT INTO countries (name, code) VALUES (?, ?)"
	result, err := d.db.Exec(insertCountryQuery, name, code)
	if err != nil {
		return Country{}, false, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Country{}, false, err
	}
	return Country{
		ID:   int(id),
		Name: name,
		Code: code,
	}, false, nil
}

func checkIfCountryExists(db *sql.DB, name string) bool {
	var id int
	getCountryQuery := "SELECT id FROM countries WHERE name = ?"
	row := db.QueryRow(getCountryQuery, name)
	err := row.Scan(&id)
	return err == nil
}

func (d *DB) RegisterUser(username string, email string, hashedPassword []byte, salt []byte) (int64, error) {
	registerNewUserQuery := "INSERT INTO users (username, email, password, salt) VALUES (?, ?, ?, ?)"
	result, err := d.db.Exec(registerNewUserQuery, username, email, hashedPassword, salt)
	if err != nil {
		return 0, fmt.Errorf("error inserting user into database: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting user ID: %w", err)
	}

	return userID, nil
}

func (d *DB) AddCovidStatistic(countryID int, date string, confirmed int, recovered int, deaths int) (int, error) {
	addCovidStatisticQuery := `
		INSERT INTO covid_statistics
		(country_id, date, confirmed, recovered, deaths)
		VALUES (?, ?, ?, ?, ?);`
	res, err := d.db.Exec(addCovidStatisticQuery, countryID, date, confirmed, recovered, deaths)
	if err != nil {
		return 0, fmt.Errorf("error inserting covid statistic into database: %w", err)
	}

	covidStatisticID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting covid statistic ID: %w", err)
	}

	return int(covidStatisticID), nil
}

func (d *DB) AddUserMonitoredCountry(userID int, countryID int) error {
	addUserMonitoredCountryQuery := `
		INSERT INTO user_monitored_countries
		(user_id, country_id)
		VALUES (?, ?);`
	_, err := d.db.Exec(addUserMonitoredCountryQuery, userID, countryID)
	if err != nil {
		return fmt.Errorf("error inserting user monitored country into database: %w", err)
	}

	return nil
}
