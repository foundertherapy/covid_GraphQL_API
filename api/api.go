package api

import (
	"covid/database"
	"covid/fetcher"
	"covid/graph"
	"covid/graph/model"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func UserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "username parameter is required", http.StatusBadRequest)
			return
		}

		var user database.User
		d := database.NewDB(db)
		user, err := d.GetUserByUsername(username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user.Salt = ""
		user.Password = ""

		user.MonitoredCountries, err = d.GetUserMonitoredCountries(user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		apiUser := MapDatabaseUserToAPIModel(&user)

		// Encode the user object as JSON and return it in the response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiUser)
	}
}

func CountryByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid country ID", http.StatusBadRequest)
			return
		}

		d := database.NewDB(db)
		country, err := d.GetCountryByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		apiCountry := MapDatabaseCountryToAPIModel(&country)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiCountry)
	}
}

func CountriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filterNameContains := r.URL.Query().Get("filterNameContains")
		filterCodeEquals := r.URL.Query().Get("filterCodeEquals")

		filter := CountryFilterInput{}
		if filterNameContains != "" {
			filter.NameContains = &filterNameContains
		}
		if filterCodeEquals != "" {
			filter.CodeEquals = &filterCodeEquals
		}

		d := database.NewDB(db)
		countries, err := d.GetCountries(nil, nil, filter.CodeEquals, filter.NameContains)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var apiCountries []*Country
		for _, country := range countries {
			apiCountries = append(apiCountries, MapDatabaseCountryToAPIModel(&country))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiCountries)
	}
}

func AddCountryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var input CountryInput
			err := json.NewDecoder(r.Body).Decode(&input)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if len(input.Code) != 2 {
				http.Error(w, "country code must be 2 characters long", http.StatusBadRequest)
				return
			}

			d := database.NewDB(db)
			country, ifExists, err := d.CreateCountry(input.Name, input.Code)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to insert new country: %v", err), http.StatusInternalServerError)
				return
			}

			if ifExists {
				http.Error(w, "country already exists", http.StatusBadRequest)
				return
			}

			url := fmt.Sprintf("/api/countries/%d", country.ID)
			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusCreated)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func UpdateCountryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			idStr := chi.URLParam(r, "id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid country ID", http.StatusBadRequest)
				return
			}

			var input CountryInput
			err = json.NewDecoder(r.Body).Decode(&input)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if len(input.Code) != 2 {
				http.Error(w, "country code must be 2 characters long", http.StatusBadRequest)
				return
			}

			d := database.NewDB(db)
			country, err := d.UpdateCountry(id, input.Name, input.Code)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to update country: %v", err), http.StatusInternalServerError)
				return
			}

			url := fmt.Sprintf("/api/countries/%d", country.ID)
			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusCreated)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func DeleteCountryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			idStr := chi.URLParam(r, "id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid country ID", http.StatusBadRequest)
				return
			}

			d := database.NewDB(db)
			err = d.DeleteCountry(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func CovidStatisticByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid CovidStatistic ID", http.StatusBadRequest)
			return
		}

		d := database.NewDB(db)
		covidStat, err := d.GetCovidStatistic(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		apiCovidStat := MapDatabaseCovidStatisticToAPIModel(&covidStat)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiCovidStat)
	}
}

func CovidStatisticsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		queryParams := r.URL.Query()
		countryID := queryParams.Get("country_id")

		countryIDInt, err := strconv.Atoi(countryID)
		if err != nil {
			http.Error(w, "Invalid country ID", http.StatusBadRequest)
			return
		}

		d := database.NewDB(db)
		covidStats, err := d.GetCovidStatistics(countryIDInt, nil, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var apiCovidStats []*CovidStatistic
		for i := range covidStats {
			apiCovidStats = append(apiCovidStats, MapDatabaseCovidStatisticToAPIModel(&covidStats[i]))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiCovidStats)
	}
}

func AddCovidStatisticHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var input CovidStatisticInput
			err := json.NewDecoder(r.Body).Decode(&input)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			countryID, err := strconv.Atoi(input.CountryID)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid country ID: %v", err), http.StatusBadRequest)
				return
			}

			date, err := time.Parse("2006-01-02", input.Date)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid date: %v", err), http.StatusBadRequest)
				return
			}

			d := database.NewDB(db)
			covidStatisticID, err := d.AddCovidStatistic(countryID, date.Format("2006-01-02"), input.Confirmed, input.Recovered, input.Deaths)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to insert new covid statistic: %v", err), http.StatusInternalServerError)
				return
			}

			url := fmt.Sprintf("/covid-stats/%d", covidStatisticID)
			w.Header().Set("Location", url)
			w.WriteHeader(http.StatusCreated)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func UpdateCovidStatisticHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		covidStatisticID, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Invalid covid statistic ID", http.StatusBadRequest)
			return
		}

		var input CovidStatisticInput
		err = json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		dateTime, err := time.Parse("2006-01-02", input.Date)
		if err != nil {
			http.Error(w, "Invalid date", http.StatusBadRequest)
			return
		}

		d := database.NewDB(db)
		_, err = d.UpdateCovidStatistic(covidStatisticID, dateTime.Format("2006-01-02"), input.Confirmed, input.Recovered, input.Deaths)
		if err != nil {
			http.Error(w, "Failed to update covid statistic", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", fmt.Sprintf("/covid-stats/%s", id))
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteCovidStatisticHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		covidStatisticID, err := strconv.Atoi(id)
		if err != nil {
			http.Error(w, "Invalid covid statistic ID", http.StatusBadRequest)
			return
		}

		d := database.NewDB(db)
		err = d.DeleteCovidStatistic(covidStatisticID)
		if err != nil {
			http.Error(w, "Failed to delete covid statistic", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", fmt.Sprintf("/api/covid-stats/%s", id))
		w.WriteHeader(http.StatusNoContent)
	}
}

func GetMonitoredCountriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userid")
		id, err := strconv.Atoi(userID)
		if err != nil {
			fmt.Println("ID: ", id)
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		d := database.NewDB(db)
		monitoredCountries, err := d.GetUserMonitoredCountries(id)
		if err != nil {
			http.Error(w, "Failed to get monitored countries", http.StatusInternalServerError)
			return
		}

		apiMonitoredCountries := MapDatabaseCountriesToAPIModels(monitoredCountries)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiMonitoredCountries)
	}
}

func AddUserMonitoredCountryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userid")
		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var input struct {
			CountryID int `json:"countryId"`
		}
		err = json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		d := database.NewDB(db)
		if err := d.AddUserMonitoredCountry(userIDInt, input.CountryID); err != nil {
			http.Error(w, "Failed to add monitored country", http.StatusInternalServerError)
			return
		}

		location := fmt.Sprintf("/users/%s/monitored-countries", userID)
		w.Header().Set("Location", location)
		w.WriteHeader(http.StatusCreated)
	}
}

func DeleteUserMonitoredCountryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userid")
		countryID := chi.URLParam(r, "countryid")
		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}
		countryIDInt, err := strconv.Atoi(countryID)
		if err != nil {
			http.Error(w, "Invalid country ID", http.StatusBadRequest)
			return
		}

		d := database.NewDB(db)
		if err := d.RemoveUserMonitoredCountry(userIDInt, countryIDInt); err != nil {
			http.Error(w, "Failed to remove monitored country", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GetTopCountriesByCaseTypeForUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse user ID from URL parameter
		userID := chi.URLParam(r, "userid")
		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		caseTypeString := chi.URLParam(r, "caseType")
		caseType := model.CaseType(strings.ToUpper(caseTypeString))
		if !caseType.IsValid() {
			http.Error(w, "Invalid case type", http.StatusBadRequest)
			return
		}

		limitString := chi.URLParam(r, "limit")
		limit, err := strconv.Atoi(limitString)
		if err != nil {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}

		d := database.NewDB(db)
		countries, err := d.GetTopCountriesByCaseTypeForUser(userIDInt, caseType.String(), limit)
		if err != nil {
			http.Error(w, "Failed to get top countries by case type", http.StatusInternalServerError)
			return
		}

		apiCountries := MapDatabaseCountriesToAPIModels(countries)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiCountries)
	}
}

func GetDeathPercentageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse country ID from URL parameter
		countryID := chi.URLParam(r, "countryId")
		countryIDInt, err := strconv.Atoi(countryID)
		if err != nil {
			http.Error(w, "Invalid country ID", http.StatusBadRequest)
			return
		}

		d := database.NewDB(db)
		deathPercentage, err := d.GetDeathPercentage(countryIDInt)
		if err != nil {
			http.Error(w, "Failed to get death percentage", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]float64{"deathPercentage": deathPercentage})
	}
}

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input := UserInput{}

		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		d := database.NewDB(db)
		err = validation(input, w, d)
		if err != nil {
			return
		}

		hashedPassword, salt, err := graph.HashPassword(input.Password)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		userID, err := d.RegisterUser(input.Username, input.Email, hashedPassword, salt)
		if err != nil {
			http.Error(w, "Failed to register user", http.StatusInternalServerError)
			return
		}

		// Generate a JWT token
		token, err := graph.GenerateToken(input.Username)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		// Return the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(LoginResponse{
			Token: token,
			User: MapDatabaseUserToAPIModel(&database.User{
				ID:       int(userID),
				Username: input.Username,
				Email:    input.Email,
			}),
		})
	}
}

func validation(input UserInput, w http.ResponseWriter, d *database.DB) error {
	if input.Username == "" || input.Email == "" || input.Password == "" {
		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
		return fmt.Errorf("username, email, and password are required")
	}

	if err := graph.ValidateUsername(input.Username); err != nil {
		return err
	}

	if err := graph.ValidateEmail(input.Email); err != nil {
		return err
	}

	if err := graph.ValidatePassword(input.Password); err != nil {
		return err
	}

	if err := d.CheckIfUserExists(input.Username); err != nil {
		return err
	}

	if err := d.CheckIfEmailExists(input.Email); err != nil {
		return err
	}

	return nil
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input UserInput

		err := json.NewDecoder(r.Body).Decode(&input)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if input.Username == "" || input.Password == "" {
			http.Error(w, "Username and password are required", http.StatusBadRequest)
			return
		}

		d := database.NewDB(db)
		user, err := d.GetUserByUsername(input.Username)
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusBadRequest)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password+user.Salt))
		if err != nil {
			http.Error(w, "Invalid username or password", http.StatusBadRequest)
			return
		}

		token, err := graph.GenerateToken(input.Username)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LoginResponse{
			Token: token,
			User:  MapDatabaseUserToAPIModel(&user),
		})
	}
}

func DeleteUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userid")
		userIDInt, err := strconv.Atoi(userID)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		d := database.NewDB(db)
		if err := d.DeleteUser(userIDInt); err != nil {
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func RefreshCovidDataForAllCountriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if err := fetcher.FetchAndUpdateData(db); err != nil {
			http.Error(w, "Failed to refresh COVID data", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
