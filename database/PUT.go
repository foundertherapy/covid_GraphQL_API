package database

import "fmt"

func (d *DB) UpdateCountry(id int, name string, code string) (Country, error) {
	updateNameCode := "UPDATE countries SET name = ?, code = ? WHERE id = ?"
	result, err := d.db.Exec(updateNameCode, name, code, id)
	if err != nil {
		return Country{}, fmt.Errorf("could not update country: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Country{}, fmt.Errorf("no countries were affected: %w", err)
	}

	if rowsAffected == 0 {
		return Country{}, fmt.Errorf("country not found")
	}

	return Country{
		ID:   id,
		Name: name,
		Code: code,
	}, nil
}

func (d *DB) UpdateCovidStatistic(id int, date string, confirmed int, recovered int, deaths int) (CovidStatistic, error) {
	updateCovidStatistic := "UPDATE covid_statistics SET date = ?, confirmed = ?, recovered = ?, deaths = ? WHERE id = ?"
	result, err := d.db.Exec(updateCovidStatistic, date, confirmed, recovered, deaths, id)
	if err != nil {
		return CovidStatistic{}, fmt.Errorf("could not update covid statistic: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return CovidStatistic{}, fmt.Errorf("no covid statistics were affected: %w", err)
	}
	if rowsAffected == 0 {
		return CovidStatistic{}, fmt.Errorf("covid statistic not found")
	}

	countryID, err := d.GetCountryIDByCovidStatisticID(id)
	if err != nil {
		return CovidStatistic{}, err
	}

	return CovidStatistic{
		ID:        id,
		Date:      date,
		Confirmed: confirmed,
		Recovered: recovered,
		Deaths:    deaths,
		CountryID: countryID,
	}, nil
}
