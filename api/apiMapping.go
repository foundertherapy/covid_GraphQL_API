package api

import (
	"covid/database"
	"fmt"
	"strconv"
)

type UserInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID                 string     `json:"id"`
	Username           string     `json:"username"`
	Email              string     `json:"email"`
	MonitoredCountries []*Country `json:"monitored_countries"`
}

type Country struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type CovidStatistic struct {
	ID        string `json:"id"`
	CountryID string `json:"country_id"`
	Date      string `json:"date"`
	Confirmed int    `json:"confirmed"`
	Recovered int    `json:"recovered"`
	Deaths    int    `json:"deaths"`
}

type CountryFilterInput struct {
	NameContains *string `json:"nameContains,omitempty"`
	CodeEquals   *string `json:"codeEquals,omitempty"`
}

type CountryInput struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type CovidStatisticInput struct {
	CountryID string `json:"countryID"`
	Date      string `json:"date"`
	Confirmed int    `json:"confirmed"`
	Recovered int    `json:"recovered"`
	Deaths    int    `json:"deaths"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

func MapDatabaseCovidStatisticsToAPIModels(covidStatistics []*database.CovidStatistic) []*CovidStatistic {
	var apiModels []*CovidStatistic
	for _, cs := range covidStatistics {
		apiModels = append(apiModels, MapDatabaseCovidStatisticToAPIModel(cs))
	}
	return apiModels
}

func MapDatabaseCovidStatisticToAPIModel(covidStatistic *database.CovidStatistic) *CovidStatistic {
	return &CovidStatistic{
		ID:        fmt.Sprint(covidStatistic.ID),
		CountryID: fmt.Sprint(covidStatistic.CountryID),
		Date:      covidStatistic.Date,
		Confirmed: covidStatistic.Confirmed,
		Recovered: covidStatistic.Recovered,
		Deaths:    covidStatistic.Deaths,
	}
}

func MapDatabaseCountryToAPIModel(country *database.Country) *Country {
	return &Country{
		ID:   fmt.Sprint(country.ID),
		Name: country.Name,
		Code: country.Code,
	}
}

func MapDatabaseCountriesToAPIModels(countries []database.Country) []*Country {
	var apiModels []*Country
	for _, country := range countries {
		apiModels = append(apiModels, MapDatabaseCountryToAPIModel(&country))
	}
	return apiModels
}

func MapDatabaseUserToAPIModel(user *database.User) *User {
	return &User{
		ID:                 fmt.Sprint(user.ID),
		Email:              user.Email,
		Username:           user.Username,
		MonitoredCountries: MapDatabaseCountriesToAPIModels(user.MonitoredCountries),
	}
}

func MapDatabaseUsersToAPIModels(users []database.User) []*User {
	var apiModels []*User
	for _, user := range users {
		apiModels = append(apiModels, MapDatabaseUserToAPIModel(&user))
	}
	return apiModels
}

func MapAPIUserToDatabaseModel(user *User) *database.User {
	return &database.User{
		ID:       mustParseInt(user.ID),
		Email:    user.Email,
		Username: user.Username,
	}
}

func MapAPICountryToDatabaseModel(country *Country) *database.Country {
	return &database.Country{
		ID:   mustParseInt(country.ID),
		Name: country.Name,
		Code: country.Code,
	}
}

func MapAPICovidStatisticToDatabaseModel(covidStat *CovidStatistic) *database.CovidStatistic {
	return &database.CovidStatistic{
		ID:        mustParseInt(covidStat.ID),
		CountryID: mustParseInt(covidStat.CountryID),
		Date:      covidStat.Date,
		Confirmed: covidStat.Confirmed,
		Recovered: covidStat.Recovered,
		Deaths:    covidStat.Deaths,
	}
}

func mustParseInt(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return i
}
