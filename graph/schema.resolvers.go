package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.27

import (
	"context"
	"covid/database"
	"covid/fetcher"
	"covid/graph/model"
	"errors"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Register is the resolver for the register field.
func (r *mutationResolver) Register(ctx context.Context, username string, email string, password string) (*model.LoginResponse, error) {
	err := ValidateUserRegistration(username, email, password, r)
	if err != nil {
		return nil, err
	}

	hashedPassword, salt, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	d := database.NewDB(r.db)
	userID, err := d.RegisterUser(username, email, hashedPassword, salt)
	if err != nil {
		return nil, err
	}

	token, err := GenerateToken(username)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token: token,
		User: &model.User{
			ID:       fmt.Sprint(userID),
			Username: username,
			Email:    email,
		},
	}, nil
}

// DeleteUser is the resolver for the deleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context, userID string) (bool, error) {
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return false, fmt.Errorf("error converting user ID %s to int: %w", userID, err)
	}

	d := database.NewDB(r.db)

	if err := d.DeleteUser(userIDInt); err != nil {
		return false, err
	}
	return true, nil
}

// AddCountry is the resolver for the addCountry field.
func (r *mutationResolver) AddCountry(ctx context.Context, input model.CountryInput) (*model.Country, error) {
	if len(input.Code) != 2 {
		return nil, errors.New("country code must be 2 characters long")
	}

	d := database.NewDB(r.db)
	country, ifExists, err := d.CreateCountry(input.Name, input.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to insert new country: %w", err)
	}

	if ifExists {
		return nil, errors.New("country already exists")
	}

	return model.MapDatabaseCountryToGQLModel(&country, &pageSize), nil
}

// UpdateCountry is the resolver for the updateCountry field.
func (r *mutationResolver) UpdateCountry(ctx context.Context, id string, name string, code string) (*model.Country, error) {
	countryID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("error converting country ID %s to int: %w", id, err)
	}
	if len(code) != 2 {
		return nil, errors.New("country code must be 2 characters long")
	}

	d := database.NewDB(r.db)
	country, err := d.UpdateCountry(countryID, name, code)
	if err != nil {
		return nil, fmt.Errorf("error updating country with ID %d: %w", countryID, err)
	}

	covidStats, err := d.GetCovidStatistics(countryID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting covid statistics for country with ID %d: %w", countryID, err)
	}
	country.CovidStatistics = covidStats

	return model.MapDatabaseCountryToGQLModel(&country, &pageSize), nil
}

// DeleteCountry is the resolver for the deleteCountry field.
func (r *mutationResolver) DeleteCountry(ctx context.Context, countryID string) (bool, error) {
	countryIDInt, err := strconv.Atoi(countryID)
	if err != nil {
		return false, fmt.Errorf("error converting country ID %s to int: %w", countryID, err)
	}

	d := database.NewDB(r.db)
	err = d.DeleteCountry(countryIDInt)
	if err != nil {
		return false, err
	}

	return true, nil
}

// AddCovidStatistic is the resolver for the addCovidStatistic field.
func (r *mutationResolver) AddCovidStatistic(ctx context.Context, input model.CovidStatisticInput) (*model.CovidStatistic, error) {
	countryID, err := strconv.Atoi(input.CountryID)
	if err != nil {
		return nil, fmt.Errorf("invalid country ID: %w", err)
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date: %w", err)
	}

	d := database.NewDB(r.db)
	covidStatisticID, err := d.AddCovidStatistic(countryID, date.Format("2006-01-02"), input.Confirmed, input.Recovered, input.Deaths)
	if err != nil {
		return nil, err
	}

	country, err := d.GetCountryByID(countryID)
	if err != nil {
		return nil, err
	}

	return &model.CovidStatistic{
		ID:        fmt.Sprint(covidStatisticID),
		Country:   model.MapDatabaseCountryToGQLModel(&country, nil),
		Date:      input.Date,
		Confirmed: input.Confirmed,
		Recovered: input.Recovered,
		Deaths:    input.Deaths,
	}, nil
}

// DeleteCovidStatistic is the resolver for the deleteCovidStatistic field.
func (r *mutationResolver) DeleteCovidStatistic(ctx context.Context, id string) (bool, error) {
	covidStatisticID, err := strconv.Atoi(id)
	if err != nil {
		return false, fmt.Errorf("invalid covid statistic ID: %w", err)
	}

	d := database.NewDB(r.db)
	err = d.DeleteCovidStatistic(covidStatisticID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// UpdateCovidStatistic is the resolver for the updateCovidStatistic field.
func (r *mutationResolver) UpdateCovidStatistic(ctx context.Context, id string, date string, confirmed int, recovered int, deaths int) (*model.CovidStatistic, error) {
	covidStatisticID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid covid statistic ID: %w", err)
	}

	dateTime, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date: %w", err)
	}

	d := database.NewDB(r.db)
	covidStatistic, err := d.UpdateCovidStatistic(covidStatisticID, dateTime.Format("2006-01-02"), confirmed, recovered, deaths)
	if err != nil {
		return nil, err
	}

	country, err := d.GetCountryByID(covidStatistic.CountryID)
	if err != nil {
		return nil, err
	}

	return &model.CovidStatistic{
		ID:        fmt.Sprint(covidStatistic.ID),
		Country:   model.MapDatabaseCountryToGQLModel(&country, nil),
		Date:      date,
		Confirmed: confirmed,
		Recovered: recovered,
		Deaths:    deaths,
	}, nil
}

// AddUserMonitoredCountry is the resolver for the addUserMonitoredCountry field.
func (r *mutationResolver) AddUserMonitoredCountry(ctx context.Context, userID string, countryID string) (*model.User, error) {
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	countryIDInt, err := strconv.Atoi(countryID)
	if err != nil {
		return nil, fmt.Errorf("invalid country ID: %w", err)
	}

	d := database.NewDB(r.db)
	if err := d.AddUserMonitoredCountry(userIDInt, countryIDInt); err != nil {
		return nil, err
	}

	var user database.User
	user, err = d.GetUserByID(userIDInt)
	if err != nil {
		return nil, err
	}

	user.MonitoredCountries, err = d.GetUserMonitoredCountries(userIDInt)
	if err != nil {
		return nil, err
	}

	return model.MapDatabaseUserToGQLModel(&user), nil
}

// RemoveUserMonitoredCountry is the resolver for the removeUserMonitoredCountry field.
func (r *mutationResolver) RemoveUserMonitoredCountry(ctx context.Context, userID string, countryID string) (*model.User, error) {
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	countryIDInt, err := strconv.Atoi(countryID)
	if err != nil {
		return nil, fmt.Errorf("invalid country ID: %w", err)
	}

	d := database.NewDB(r.db)
	if err := d.RemoveUserMonitoredCountry(userIDInt, countryIDInt); err != nil {
		return nil, err
	}

	var user database.User
	user, err = d.GetUserByID(userIDInt)
	if err != nil {
		return nil, err
	}

	user.MonitoredCountries, err = d.GetUserMonitoredCountries(userIDInt)
	if err != nil {
		return nil, err
	}

	return model.MapDatabaseUserToGQLModel(&user), nil
}

// RefreshCovidDataForAllCountries is the resolver for the refreshCovidDataForAllCountries field.
func (r *mutationResolver) RefreshCovidDataForAllCountries(ctx context.Context) (bool, error) {
	if err := fetcher.FetchAndUpdateData(r.db); err != nil {
		return false, err
	}
	return true, nil
}

// Login is the resolver for the login field.
func (r *queryResolver) Login(ctx context.Context, username string, password string) (*model.LoginResponse, error) {
	d := database.NewDB(r.db)
	user, err := d.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.MonitoredCountries, err = d.GetUserMonitoredCountries(user.ID)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password+user.Salt))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	token, err := GenerateToken(username)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token: token,
		User:  model.MapDatabaseUserToGQLModel(&user),
	}, nil
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, username *string, email *string) (*model.User, error) {
	if username == nil && email == nil {
		return nil, errors.New("username or email must be provided")
	}

	var user database.User
	var err error
	d := database.NewDB(r.db)
	if username != nil {
		user, err = d.GetUserByUsername(*username)
	} else if email != nil {
		user, err = d.GetUserByEmail(*email)
	}

	if err != nil {
		return nil, err
	}
	user.Salt = ""
	user.Password = ""

	user.MonitoredCountries, err = d.GetUserMonitoredCountries(user.ID)
	if err != nil {
		return nil, err
	}

	return model.MapDatabaseUserToGQLModel(&user), nil
}

// Country is the resolver for the country field.
func (r *queryResolver) Country(ctx context.Context, id string) (*model.Country, error) {
	countryIDInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid country ID: %w", err)
	}

	d := database.NewDB(r.db)
	country, err := d.GetCountryByID(countryIDInt)
	if err != nil {
		return nil, err
	}

	return model.MapDatabaseCountryToGQLModel(&country, &pageSize), nil
}

// Countries is the resolver for the countries field.
func (r *queryResolver) Countries(ctx context.Context, first *int, after *string, filter *model.CountryFilterInput) (*model.CountriesConnection, error) {
	var codeEquals, nameContains *string
	if filter != nil {
		codeEquals = filter.CodeEquals
		nameContains = filter.NameContains
	}

	d := database.NewDB(r.db)
	countries, err := d.GetCountries(first, after, codeEquals, nameContains)
	if err != nil {
		return nil, err
	}

	//loop over each country and get the covid stats
	for i := range countries {
		covidStats, err := d.GetCovidStatistics(countries[i].ID, nil, nil)
		if err != nil {
			return nil, err
		}
		countries[i].CovidStatistics = covidStats
	}

	return model.CreateMapDatabaseCountriesToConnection(countries, &pageSize), nil
}

// MonitoredCountries is the resolver for the monitoredCountries field.
func (r *queryResolver) MonitoredCountries(ctx context.Context, userID string) ([]*model.Country, error) {
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	d := database.NewDB(r.db)
	countries, err := d.GetUserMonitoredCountries(userIDInt)
	if err != nil {
		return nil, err
	}

	//loop over each country and get the covid stats
	for i := range countries {
		covidStats, err := d.GetCovidStatistics(countries[i].ID, nil, nil)
		if err != nil {
			return nil, err
		}
		countries[i].CovidStatistics = covidStats
	}

	return model.MapDatabaseCountriesToGQLModels(countries), nil
}

// CovidStatistics is the resolver for the covidStatistics field.
func (r *queryResolver) CovidStatistics(ctx context.Context, countryID string, after *string, first *int) (*model.CovidStatisticConnection, error) {
	countryIDInt, err := strconv.Atoi(countryID)
	if err != nil {
		return nil, fmt.Errorf("invalid country ID %w", err)
	}

	d := database.NewDB(r.db)
	covidStats, err := d.GetCovidStatistics(countryIDInt, first, after)
	if err != nil {
		return nil, err
	}

	return model.CreateMapDatabaseCovidStatsToConnection(covidStats, pageSize), nil
}

// CovidStatistic is the resolver for the covidStatistic field.
func (r *queryResolver) CovidStatistic(ctx context.Context, id string) (*model.CovidStatistic, error) {

	IDInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid covid statistic ID: %w", err)
	}
	d := database.NewDB(r.db)
	covidStat, err := d.GetCovidStatistic(IDInt)
	if err != nil {
		return nil, err
	}

	//get the country for the covid stat
	country, err := d.GetCountryByID(covidStat.CountryID)
	if err != nil {
		return nil, err
	}
	covidStat.Country = country

	return model.MapDatabaseCovidStatisticToGQLModel(&covidStat), nil
}

// DeathPercentage is the resolver for the deathPercentage field.
func (r *queryResolver) DeathPercentage(ctx context.Context, countryID string) (float64, error) {
	countryIDInt, err := strconv.Atoi(countryID)
	if err != nil {
		return 0, fmt.Errorf("invalid country ID: %w", err)
	}

	d := database.NewDB(r.db)
	return d.GetDeathPercentage(countryIDInt)
}

// TopCountriesByCaseTypeForUser is the resolver for the topCountriesByCaseTypeForUser field.
func (r *queryResolver) TopCountriesByCaseTypeForUser(ctx context.Context, caseType model.CaseType, limit int, userID string) ([]*model.Country, error) {
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	d := database.NewDB(r.db)
	countries, err := d.GetTopCountriesByCaseTypeForUser(userIDInt, caseType.String(), limit)
	if err != nil {
		return nil, err
	}

	//loop over each country and get the covid stats
	for i := range countries {
		covidStats, err := d.GetCovidStatistics(countries[i].ID, nil, nil)
		if err != nil {
			return nil, err
		}
		countries[i].CovidStatistics = covidStats
	}

	return model.MapDatabaseCountriesToGQLModels(countries), nil
}

// CovidStatisticUpdated is the resolver for the covidStatisticUpdated field.
func (r *subscriptionResolver) CovidStatisticUpdated(ctx context.Context, countryIDs []string) (<-chan []*model.CovidStatistic, error) {
	var countryIDsInt []int
	for _, id := range countryIDs {
		countryIDInt, err := strconv.Atoi(id)
		if err != nil {
			return nil, fmt.Errorf("invalid country ID: %w", err)
		}
		countryIDsInt = append(countryIDsInt, countryIDInt)
	}

	updatedCovidStats := make(chan []*model.CovidStatistic)
	d := database.NewDB(r.db)
	go updateCovidStatsRoutine(d, countryIDsInt, updatedCovidStats)
	return updatedCovidStats, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
