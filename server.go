package main

import (
	"covid/api"
	"covid/database"
	"covid/fetcher"
	"covid/graph"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

func authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bypass the middleware for the login request
		if r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authorizationHeader, "Bearer ")
		_, err := graph.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	fetcher.StartFetchingRoutine(db, 24*time.Hour)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	r := graph.NewResolver(db)
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: r}))

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Handle("/", playground.Handler("GraphQL playground", "/login"))
	router.Handle("/login", srv)

	router.Group(func(r chi.Router) {
		// r.Use(authenticationMiddleware)
		r.Handle("/query", srv)
		r.HandleFunc("/api/user", api.UserHandler(db))
		r.HandleFunc("/api/countries", api.CountriesHandler(db))
		r.HandleFunc("/api/countries/create", api.AddCountryHandler(db))
		r.HandleFunc("/api/countries/{id}/update", api.UpdateCountryHandler(db))
		r.HandleFunc("/api/countries/{id}/delete", api.DeleteCountryHandler(db))
		r.HandleFunc("/api/countries/{id}", api.CountryByIDHandler(db))
		r.HandleFunc("/api/covid-stats/{id}", api.CovidStatisticByIDHandler(db))
		r.HandleFunc("/api/covid-stats", api.CovidStatisticsHandler(db))
		r.HandleFunc("/api/covid-stats/create", api.AddCovidStatisticHandler(db))
		r.HandleFunc("/api/covid-stats/{id}", api.UpdateCovidStatisticHandler(db))
		r.HandleFunc("/api/covid-stats/{id}", api.DeleteCovidStatisticHandler(db))
		r.HandleFunc("/api/users/{userid}/monitored-countries", api.GetMonitoredCountriesHandler(db))
		r.HandleFunc("/api/users/{userid}/monitored-countries", api.AddUserMonitoredCountryHandler(db))
		r.HandleFunc("/api/users/{userid}/monitored-countries/{countryid}", api.DeleteUserMonitoredCountryHandler(db))
		r.HandleFunc("/api/countries/top-by-case-type/{caseType}/{limit}/{userid}", api.GetTopCountriesByCaseTypeForUserHandler(db))
		r.HandleFunc("/api/countries/{countryId}/death-percentage", api.GetDeathPercentageHandler(db))
		r.HandleFunc("/api/register-api", api.RegisterHandler(db))
		r.HandleFunc("/api/login-api", api.LoginHandler(db))
		r.HandleFunc("/api/users/{userid}", api.DeleteUserHandler(db))
		r.HandleFunc("/api/refresh-covid-data", api.RefreshCovidDataForAllCountriesHandler(db))

	})

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
