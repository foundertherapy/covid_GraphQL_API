package model

import (
	"covid/database"
	"encoding/base64"
	"fmt"
	"strconv"
)

func CreateMapDatabaseCovidStatsToConnection(covidStats []database.CovidStatistic, pageSize int) *CovidStatisticConnection {
	var edges []*CovidStatisticEdge
	// truncate first before anything to ensure the correct number of records are returned
	if pageSize > 0 && len(covidStats) > pageSize {
		covidStats = covidStats[:pageSize] // Truncate the extra records
	}

	gqlModelsCovidStats := MapDatabaseCovidStatsToGQLModel(covidStats)

	for _, covidStat := range gqlModelsCovidStats {
		edge := &CovidStatisticEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(covidStat.ID)),
			Node:   covidStat,
		}
		edges = append(edges, edge)
	}

	var endCursor *string
	hasNextPage := false
	if len(covidStats) == pageSize {
		//because we have truncated one record, we can be sure that there are more records
		hasNextPage = true
	}

	if len(covidStats) > 0 {
		endCursorValue := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(covidStats[len(covidStats)-1].ID)))
		endCursor = &endCursorValue
	}

	return &CovidStatisticConnection{
		PageInfo: &PageInfo{
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		},
		Edges: edges,
	}
}

func MapDatabaseCovidStatsToGQLModel(covidStats []database.CovidStatistic) []*CovidStatistic {
	var gqlModels []*CovidStatistic
	for _, covidStat := range covidStats {
		gqlModels = append(gqlModels, MapDatabaseCovidStatisticToGQLModel(&covidStat))
	}
	return gqlModels
}

func MapDatabaseCovidStatisticToGQLModel(covidStatistic *database.CovidStatistic) *CovidStatistic {
	return &CovidStatistic{
		ID:        fmt.Sprint(covidStatistic.ID),
		Country:   MapDatabaseCountryToGQLModel(&covidStatistic.Country, nil),
		Date:      covidStatistic.Date,
		Confirmed: covidStatistic.Confirmed,
		Recovered: covidStatistic.Recovered,
		Deaths:    covidStatistic.Deaths,
	}
}

func MapDatabaseCountryToGQLModel(country *database.Country, pageSize *int) *Country {
	// If pageSize is nil or 0, do not paginate
	if pageSize == nil || *pageSize == 0 {
		return &Country{
			ID:         fmt.Sprint(country.ID),
			Name:       country.Name,
			Code:       country.Code,
			CovidStats: CreateMapDatabaseCovidStatsToConnection(country.CovidStatistics, 0),
		}
	}

	return &Country{
		ID:         fmt.Sprint(country.ID),
		Name:       country.Name,
		Code:       country.Code,
		CovidStats: CreateMapDatabaseCovidStatsToConnection(country.CovidStatistics, *pageSize),
	}
}

func MapDatabaseCovidStatisticsToGQLModels(covidStatistics []*database.CovidStatistic) []*CovidStatistic {
	var gqlModels []*CovidStatistic
	for _, cs := range covidStatistics {
		gqlModels = append(gqlModels, MapDatabaseCovidStatisticToGQLModel(cs))
	}
	return gqlModels
}

func MapDatabaseUserToGQLModel(user *database.User) *User {
	return &User{
		ID:                 fmt.Sprint(user.ID),
		Email:              user.Email,
		Username:           user.Username,
		MonitoredCountries: MapDatabaseCountriesToGQLModels(user.MonitoredCountries),
	}
}

func MapDatabaseCountriesToGQLModels(countries []database.Country) []*Country {
	var gqlModels []*Country
	for _, country := range countries {
		gqlModels = append(gqlModels, MapDatabaseCountryToGQLModel(&country, nil))
	}
	return gqlModels
}

func CreateMapDatabaseCountriesToConnection(countries []database.Country, pageSize *int) *CountriesConnection {
	var edges []*CountryEdge
	if *pageSize > 0 && len(countries) > *pageSize {
		countries = countries[:*pageSize]
	}

	gqlModelsCountries := MapDatabaseCountriesToGQLModels(countries)

	for _, country := range gqlModelsCountries {
		edge := &CountryEdge{
			Cursor: base64.StdEncoding.EncodeToString([]byte(country.ID)),
			Node:   country,
		}
		edges = append(edges, edge)
	}

	var endCursor *string
	hasNextPage := false
	if len(countries) == *pageSize {
		hasNextPage = true
	}

	if len(countries) > 0 {
		endCursorValue := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(countries[len(countries)-1].ID)))
		endCursor = &endCursorValue
	}

	return &CountriesConnection{
		PageInfo: &PageInfo{
			EndCursor:   endCursor,
			HasNextPage: hasNextPage,
		},
		Edges: edges,
	}
}
