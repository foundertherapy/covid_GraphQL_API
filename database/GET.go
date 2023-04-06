package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

func (d *DB) GetUserByID(id int) (User, error) {
	user := User{}
	getUserQuery := "SELECT id, username, email, password FROM users WHERE id = ?"
	row := d.db.QueryRow(getUserQuery, id)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		return user, fmt.Errorf("could not get user: %w", err)
	}
	return user, nil
}

func (d *DB) GetUserByUsername(username string) (User, error) {
	user := User{}
	getUserQuery := "SELECT id, username, email, password, salt FROM users WHERE username = ?"
	row := d.db.QueryRow(getUserQuery, username)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt)
	if err != nil {
		return user, fmt.Errorf("could not get user: %w", err)
	}
	return user, nil
}

// get user by email:
func (d *DB) GetUserByEmail(email string) (User, error) {
	user := User{}
	getUserQuery := "SELECT id, username, email, password, salt FROM users WHERE email = ?"
	row := d.db.QueryRow(getUserQuery, email)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt)
	if err != nil {
		return user, fmt.Errorf("could not get user: %w", err)
	}
	return user, nil
}

func (d *DB) CheckIfUserExists(username string) error {
	var count int
	countUsers := "SELECT COUNT(*) FROM users WHERE username = ?"
	row := d.db.QueryRow(countUsers, username)
	err := row.Scan(&count)
	if err != nil {
		return fmt.Errorf("error while scanning number of users: %w", err)
	}
	if count > 0 {
		return errors.New("username already exists")
	}
	return nil
}

func (d *DB) CheckIfEmailExists(email string) error {
	var count int
	countUsersEmails := "SELECT COUNT(*) FROM users WHERE email = ?"
	row := d.db.QueryRow(countUsersEmails, email)
	err := row.Scan(&count)
	if err != nil {
		return fmt.Errorf("error while scanning number of users: %w", err)
	}

	if count > 0 {
		return errors.New("email already exists")
	}
	return nil
}

// get one sinle covid statistic
func (d *DB) GetCovidStatistic(id int) (CovidStatistic, error) {
	covidStatistic := CovidStatistic{}
	getCovidStatisticQuery := "SELECT id, country_id, date, confirmed, recovered, deaths FROM covid_statistics WHERE id = ?"
	row := d.db.QueryRow(getCovidStatisticQuery, id)
	err := row.Scan(&covidStatistic.ID, &covidStatistic.CountryID, &covidStatistic.Date, &covidStatistic.Confirmed, &covidStatistic.Recovered, &covidStatistic.Deaths)
	if err != nil {
		return covidStatistic, fmt.Errorf("could not get covid statistic: %w", err)
	}
	return covidStatistic, nil
}

type covidStatisticsQuery struct {
	sqlQuery string
	args     []any
}

func (d *DB) GetCovidStatistics(countryID int, first *int, after *string) ([]CovidStatistic, error) {
	covidStatsQuery, err := buildCovidStatisticsQuery(countryID, after, first)
	if err != nil {
		return nil, err
	}

	rows, err := d.db.Query(covidStatsQuery.sqlQuery, covidStatsQuery.args...)
	if err != nil {
		return nil, fmt.Errorf("could not get covid statistics for country: %w", err)
	}
	defer rows.Close()

	var covidStatistics []CovidStatistic
	for rows.Next() {
		err := mapCovidStatisticsAndCountryFromRows(rows, &covidStatistics)
		if err != nil {
			return nil, err
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return covidStatistics, nil
}

func buildCovidStatisticsQuery(countryID int, after *string, first *int) (covidStatisticsQuery, error) {
	query := covidStatisticsQuery{
		sqlQuery: `
			SELECT cs.id, cs.date, cs.confirmed, cs.deaths, cs.recovered, c.id, c.name, c.code
			FROM covid_statistics cs
			JOIN countries c ON c.id = cs.country_id
			WHERE c.id = ?`,
		args: []any{countryID},
	}

	// add pagination:
	if after != nil {
		cursor, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return query, err
		}
		query.sqlQuery += " AND cs.id > ?"
		query.args = append(query.args, string(cursor))
	}

	//get one mroe record to check if there is a next page later:
	if first != nil {
		query.sqlQuery += " LIMIT ?"
		query.args = append(query.args, *first+1)
	}

	return query, nil
}

func mapCovidStatisticsAndCountryFromRows(rows *sql.Rows, covidStatistics *[]CovidStatistic) error {
	covidStatistic := CovidStatistic{}
	country := Country{}

	err := rows.Scan(
		&covidStatistic.ID, &covidStatistic.Date, &covidStatistic.Confirmed,
		&covidStatistic.Deaths, &covidStatistic.Recovered,
		&country.ID, &country.Name, &country.Code,
	)
	if err != nil {
		return fmt.Errorf("could not scan covid statistic: %w", err)
	}
	covidStatistic.Date = covidStatistic.Date[:10]

	covidStatistic.Country = country
	*covidStatistics = append(*covidStatistics, covidStatistic)
	return nil
}

// get a speicific country by its id:
func (d *DB) GetCountryByID(id int) (Country, error) {
	country := Country{}
	getCountryQuery := "SELECT id, name, code FROM countries WHERE id = ?"
	row := d.db.QueryRow(getCountryQuery, id)

	CovidStatistic, err := d.GetCovidStatistics(id, nil, nil)
	if err != nil {
		return country, fmt.Errorf("could not get covid statistics for country: %s", err)
	}
	country.CovidStatistics = CovidStatistic
	if row.Scan(&country.ID, &country.Name, &country.Code) != nil {
		return country, fmt.Errorf("could not scan country row: %w", err)
	}
	return country, nil
}

func (d *DB) GetCountryIDByCovidStatisticID(covidStatisticID int) (int, error) {
	getCountryIDQuery := "SELECT country_id FROM covid_statistics WHERE id = ?"
	var countryID int
	err := d.db.QueryRow(getCountryIDQuery, covidStatisticID).Scan(&countryID)
	if err != nil {
		return 0, fmt.Errorf("error getting country ID for covid statistic with ID %d: %w", covidStatisticID, err)
	}
	return countryID, nil
}

func (d *DB) GetUserMonitoredCountries(userID int) ([]Country, error) {
	getMonitoredCountriesQuery := `
		SELECT c.id, c.name, c.code
		FROM user_monitored_countries umc
		JOIN countries c ON c.id = umc.country_id
		WHERE umc.user_id = ?`
	rows, err := d.db.Query(getMonitoredCountriesQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get monitored countries: %w", err)
	}
	defer rows.Close()

	var MonitoredCountries []Country
	for rows.Next() {
		country := Country{}
		err := rows.Scan(&country.ID, &country.Name, &country.Code)
		if err != nil {
			return MonitoredCountries, fmt.Errorf("could not scan monitored country: %w", err)
		}
		MonitoredCountries = append(MonitoredCountries, country)
	}

	if err := rows.Err(); err != nil {
		return MonitoredCountries, fmt.Errorf("error with rows: %w", err)
	}
	return MonitoredCountries, nil
}

type countriesQuery struct {
	sqlQuery string
	args     []any
}

func (d *DB) GetCountries(first *int, after *string, CodeEquals *string, nameContains *string) ([]Country, error) {
	countriesQuery, err := buildCountriesQuery(after, first, CodeEquals, nameContains)
	if err != nil {
		return nil, err
	}
	rows, err := d.db.Query(countriesQuery.sqlQuery, countriesQuery.args...)
	if err != nil {
		return nil, fmt.Errorf("could not get countries: %w", err)
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		err := mapCountryFromRows(rows, &countries)
		if err != nil {
			return nil, err
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return countries, nil
}

func buildCountriesQuery(after *string, first *int, CodeEquals *string, nameContains *string) (countriesQuery, error) {
	query := countriesQuery{
		sqlQuery: `
			SELECT id, name, code
			FROM countries`,
	}
	var conditions []string

	// Add code equals condition:
	if CodeEquals != nil {
		conditions = append(conditions, "code = ?")
		query.args = append(query.args, *CodeEquals)
	}

	// Add name contains condition:
	if nameContains != nil {
		conditions = append(conditions, "name LIKE ?")
		query.args = append(query.args, "%"+*nameContains+"%")
	}

	// Add pagination:
	if after != nil {
		cursor, err := base64.StdEncoding.DecodeString(*after)
		if err != nil {
			return query, err
		}
		conditions = append(conditions, "id > ?")
		query.args = append(query.args, string(cursor))
	}

	if len(conditions) > 0 {
		query.sqlQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Get one more record to check if there is a next page later:
	if first != nil {
		query.sqlQuery += " LIMIT ?"
		query.args = append(query.args, *first+1)
	}

	return query, nil
}

func mapCountryFromRows(rows *sql.Rows, countries *[]Country) error {
	country := Country{}
	err := rows.Scan(&country.ID, &country.Name, &country.Code)
	if err != nil {
		return fmt.Errorf("could not scan country: %w", err)
	}
	*countries = append(*countries, country)
	return nil
}

// get percentage of deaths in a country:
func (d *DB) GetDeathPercentage(countryID int) (float64, error) {
	//case the Sum of deaths to a real number instead of an integer.
	// Cause Otherwise the result will be 0 cause the division of two integers is an integer.
	getDeathPercentageQuery := `
		SELECT SUM(deaths) * 1.0 / SUM(confirmed) * 100
		FROM covid_statistics
		WHERE country_id = ?`
	var deathPercentage float64
	err := d.db.QueryRow(getDeathPercentageQuery, countryID).Scan(&deathPercentage)
	if err != nil {
		return 0, fmt.Errorf("could not get death percentage: %w", err)
	}

	fmt.Println("Death percentage:", deathPercentage)
	return deathPercentage, nil
}

func (d *DB) GetTopCountriesByCaseTypeForUser(userID int, caseType string, limit int) ([]Country, error) {
	getTopCountriesByCaseTypeForUserQuery := buildTopCountriesByCaseTypeForUserQuery(userID, caseType, limit)
	rows, err := d.db.Query(getTopCountriesByCaseTypeForUserQuery, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("could not get top countries by case type for user: %w", err)
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		err := mapCountryFromRows(rows, &countries)
		if err != nil {
			return nil, fmt.Errorf("could not map country from rows: %w", err)
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("error with rows: %w", err)
	}

	return countries, nil
}

func buildTopCountriesByCaseTypeForUserQuery(userID int, caseType string, limit int) string {
	getTopCountriesByCaseTypeForUserQuery := `
		SELECT c.id, c.name, c.code
		FROM covid_statistics cs
		JOIN countries c ON c.id = cs.country_id
		JOIN (
			SELECT country_id, MAX(date) as latest_date
			FROM covid_statistics
			GROUP BY country_id
		) latest_stats ON cs.country_id = latest_stats.country_id AND cs.date = latest_stats.latest_date
		WHERE cs.country_id IN (
			SELECT country_id
			FROM user_monitored_countries
			WHERE user_id = ?
		)
		ORDER BY cs.` + caseType + ` DESC
		LIMIT ?`

	return getTopCountriesByCaseTypeForUserQuery
}

func (d *DB) GetLatestCovidStatisticsByCountryID(countryID int) (CovidStatistic, error) {
	getLatestCovidStatisticsByCountryIDQuery := `
		SELECT id, country_id, confirmed, deaths, recovered, date
		FROM covid_statistics
		WHERE country_id = ?
		ORDER BY date DESC
		LIMIT 1`
	row := d.db.QueryRow(getLatestCovidStatisticsByCountryIDQuery, countryID)
	covidStatistics := CovidStatistic{}
	err := row.Scan(&covidStatistics.ID, &covidStatistics.CountryID, &covidStatistics.Confirmed, &covidStatistics.Deaths, &covidStatistics.Recovered, &covidStatistics.Date)
	if err != nil {
		return covidStatistics, fmt.Errorf("could not get latest covid statistics by country id: %w", err)
	}
	return covidStatistics, nil
}

func (d *DB) CheckCovidStatisticExists(countryID int, date string) (bool, error) {
	checkCovidStatisticExistsQuery := `
		SELECT EXISTS (
			SELECT 1
			FROM covid_statistics
			WHERE country_id = ? AND date = ?
		)`
	var exists bool
	err := d.db.QueryRow(checkCovidStatisticExistsQuery, countryID, date).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("could not check covid statistic exists: %w", err)
	}
	return exists, nil
}
