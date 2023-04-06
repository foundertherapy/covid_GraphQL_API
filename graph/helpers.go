package graph

import (
	"covid/database"
	"covid/graph/model"
	"crypto/rand"
	"database/sql"
	"errors"
	"log"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var pageSize int = 2

type Resolver struct {
	db *sql.DB
}

func NewResolver(db *sql.DB) *Resolver {
	return &Resolver{db: db}
}

type PageInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor,omitempty"`
}

type CountryEdge struct {
	Cursor string           `json:"cursor"`
	Node   database.Country `json:"node"`
}

type CountryConnection struct {
	PageInfo *PageInfo      `json:"pageInfo"`
	Edges    []*CountryEdge `json:"edges"`
}

func ValidateUserRegistration(username string, email string, password string, r *mutationResolver) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}

	if err := ValidateEmail(email); err != nil {
		return err
	}

	if err := ValidatePassword(password); err != nil {
		return err
	}
	d := database.NewDB(r.db)

	if err := d.CheckIfUserExists(username); err != nil {
		return err
	}

	if err := d.CheckIfEmailExists(email); err != nil {
		return err
	}

	return nil
}

func ValidateUsername(username string) error {
	const minUsernameLength = 3
	var usernameRegex = regexp.MustCompile(`^[a-z]+(-[a-z]+)*$`)

	if len(username) < minUsernameLength {
		return errors.New("username must be at least 3 characters long")
	}

	if !usernameRegex.MatchString(username) {
		return errors.New("username must be lowercase and can only contain hyphens")
	}

	return nil
}

func ValidateEmail(email string) error {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	if !emailRegex.MatchString(email) {
		return errors.New("invalid email address")
	}

	return nil
}

func ValidatePassword(password string) error {
	const minPasswordLength = 8

	// As Go doesn't have lookbehind or lookahead, we need to break the validation
	// into separate checks:
	var lowerRegex = regexp.MustCompile(`[a-z]`)
	var upperRegex = regexp.MustCompile(`[A-Z]`)
	var digitRegex = regexp.MustCompile(`\d`)
	var specialRegex = regexp.MustCompile(`[@$!%*#?&]`)

	if len(password) < minPasswordLength {
		return errors.New("password must be at least 8 characters long")
	}

	if !lowerRegex.MatchString(password) || !upperRegex.MatchString(password) || !digitRegex.MatchString(password) || !specialRegex.MatchString(password) {
		return errors.New("password must be strong and contain at least 1 uppercase, 1 lowercase, 1 number, and 1 special character")
	}

	return nil
}

func HashPassword(password string) ([]byte, []byte, error) {
	salt := make([]byte, 16)
	// fill the salt byte slice with random bytes
	_, err := rand.Read(salt)
	if err != nil {
		return nil, nil, err
	}

	// use blowfish algorithm  to hash the password as it is fster than AES
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+string(salt)), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}
	return hashedPassword, salt, nil
}

func updateCovidStatsRoutine(d *database.DB, countryIDsInt []int, updatedCovidStats chan []*model.CovidStatistic) {
	ticker := time.NewTicker(24 * time.Second) // Check for updates daily
	for range ticker.C {
		var newCovidStats []*database.CovidStatistic
		for _, id := range countryIDsInt {
			covidStat, err := d.GetLatestCovidStatisticsByCountryID(id)
			if err != nil {
				log.Printf("Error fetching latest Covid statistic for country %d: %v", id, err)
				continue
			}
			country, err := d.GetCountryByID(covidStat.CountryID)
			if err != nil {
				log.Printf("Error fetching country %d: %v", covidStat.CountryID, err)
				continue
			}
			covidStat.Country = country
			newCovidStats = append(newCovidStats, &covidStat)
		}
		updatedCovidStats <- model.MapDatabaseCovidStatisticsToGQLModels(newCovidStats)
	}
}
