package fetcher

import (
	"covid/database"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Covid19APIResponse struct {
	Country     string `json:"Country"`
	CountryCode string `json:"CountryCode"`
	Province    string `json:"Province"`
	Cases       int    `json:"Cases"`
	Status      string `json:"Status"`
	Date        string `json:"Date"`
}

var counter int = 0

func StartFetchingRoutine(db *sql.DB, updateInterval time.Duration) {
	go func() {
		FetchAndUpdateData(db) // Call this function immediately for the initial data fetch.
		for range time.Tick(updateInterval) {
			FetchAndUpdateData(db)
		}
	}()
}

func FetchAndUpdateData(db *sql.DB) error {
	d := database.NewDB(db)
	countries, err := d.GetCountries(nil, nil, nil, nil)
	if err != nil {
		log.Printf("Error fetching country list: %v", err)
		return err
	}

	for _, country := range countries {
		confirmedData, err := FetchDailyDataForCountry(country.Name, "confirmed")
		if err != nil {
			log.Printf("Error fetching confirmed data for country %s: %v", country.Name, err)
			continue
		}

		deathsData, err := FetchDailyDataForCountry(country.Name, "deaths")
		if err != nil {
			log.Printf("Error fetching deaths data for country %s: %v", country.Name, err)
			continue
		}

		recoveredData, err := FetchDailyDataForCountry(country.Name, "recovered")
		if err != nil {
			log.Printf("Error fetching recovered data for country %s: %v", country.Name, err)
			continue
		}

		if err := UpdateCountryData(db, country.Name, country.ID, confirmedData, deathsData, recoveredData); err != nil {
			log.Printf("Error updating country data for %s: %v", country.Name, err)
			continue
		}
	}
	return nil
}

func FetchDailyDataForCountry(countryName string, status string) ([]Covid19APIResponse, error) {
	url := fmt.Sprintf("https://api.covid19api.com/dayone/country/%s/status/%s", countryName, status)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = handleRateLimitingError(resp, countryName, status)
	if err != nil {
		return nil, err
	}

	var data []Covid19APIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func handleRateLimitingError(resp *http.Response, countryName string, status string) error {
	// Sleep for 2 seconds to avoid rate limiting by the API caused by too many requests.
	time.Sleep(5 * time.Second)

	if resp.StatusCode == http.StatusTooManyRequests {
		log.Println("Too many requests, retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
		FetchDailyDataForCountry(countryName, status)
		if counter > 5 {
			counter = 0
			return errors.New("too many retries, aborting")
		}
		counter++
	}
	return nil
}

// FindCountryByName retrieves a country record from the database by name.
func FindCountryByName(db *sql.DB, name string) (database.Country, error) {
	var country database.Country
	err := db.QueryRow("SELECT id, name, code FROM countries WHERE name = ?", name).Scan(&country.ID, &country.Name, &country.Code)
	if err != nil {
		return database.Country{}, err
	}
	return country, nil
}

// UpdateCountryData updates covid statistics for a specific country in the database.
func UpdateCountryData(db *sql.DB, countryName string, countryID int, confirmedData, deathsData, recoveredData []Covid19APIResponse) error {
	d := database.NewDB(db)

	country, err := d.GetCountryByID(countryID)
	if err != nil {
		return err
	}

	for i := 0; i < len(confirmedData); i++ {
		date, err := time.Parse(time.RFC3339, confirmedData[i].Date)
		if err != nil {
			return err
		}
		dateStr := date.Format("2006-01-02")

		exists, err := d.CheckCovidStatisticExists(country.ID, dateStr)
		if err != nil {
			return err
		}

		if !exists {
			_, err := d.AddCovidStatistic(country.ID, dateStr, confirmedData[i].Cases, recoveredData[i].Cases, deathsData[i].Cases)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
