package database

import (
	"fmt"
)

func (d *DB) DeleteCovidStatistic(id int) error {
	deleteCovidStatistic := "DELETE FROM covid_statistics WHERE id = ?"
	result, err := d.db.Exec(deleteCovidStatistic, id)
	if err != nil {
		return fmt.Errorf("could not delete covid statistic: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("no covid stats were affected: %s", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("covid stat not found")
	}

	return nil

}

func (d *DB) DeleteCountry(countryID int) error {
	deleteCountry := "DELETE FROM countries WHERE id = ?"
	result, err := d.db.Exec(deleteCountry, countryID)
	if err != nil {
		return fmt.Errorf("could not delete country: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("no countries were affected: %s", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("country not found")
	}

	return nil
}

func (d *DB) DeleteUser(id int) error {
	deleteUserQuery := "DELETE FROM users WHERE id = ?"
	res, err := d.db.Exec(deleteUserQuery, id)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get number of affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %d not found", id)
	}

	return nil
}

func (d *DB) RemoveUserMonitoredCountry(userID int, countryID int) error {
	removeUserMonitoredCountryQuery := `
		DELETE FROM user_monitored_countries
		WHERE user_id = ? AND country_id = ?;`
	result, err := d.db.Exec(removeUserMonitoredCountryQuery, userID, countryID)
	if err != nil {
		return fmt.Errorf("error removing user monitored country from database: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if affectedRows == 0 {
		return fmt.Errorf("no rows were affected by the delete statement")
	}
	return nil
}
