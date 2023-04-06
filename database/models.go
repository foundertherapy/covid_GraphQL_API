package database

type Country struct {
	ID              int
	Name            string
	Code            string
	CovidStatistics []CovidStatistic
}

type CovidStatistic struct {
	ID        int
	CountryID int
	Country   Country
	Date      string
	Confirmed int
	Recovered int
	Deaths    int
}

type User struct {
	ID                 int
	Username           string
	Email              string
	Password           string
	Salt               string
	MonitoredCountries []Country
}
