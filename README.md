# covid_GraphQL_API

## This project contains both GraphQL API and a REST API for fetching data from `https://api.covid19api.com` for certain countries.

### The features that are included are the following:
* Registering and Loggin in for Users.
* Additional layer of authintication using JWT Tokens
* Adding, Removing, Updating, and deleting countries
* Adding, Removing, Updating, and deleteing indivisual Covid Statistics.
* Pagination for GraphQL Responses
* Monitoring certain countrie for certain users
* Subscription to certain countries and getting updates on stats for them daily

### To run the server, you'll need to do the following:
1. Clone the repo
2. Run `go mod tidy` to get all of the reuquired packages from "go.mod".
3. Change the port inside of `server.go` file if needed, by default it runs on ```8080```
4. Run `go run server.go` in the terminal to launch the server

## To use the GraphQL UI to see all of the documentations for each query, head to `http://localhost:8080/` to see the UI. However, to interact with the API, you'll need to use `http://localhost:8080/query`  
*Note: Keep in mind you'll need to provide authorization when using this approach to send requests.  

## To use the REST API, you can use the following URLs to query the API by navigating to /api/
- GET /user: Returns a User by username.
- GET /countries/{id}: Returns a Country by ID.
- GET /countries: Returns a list of countries.
- POST /countries: Creates a new Country.
- PUT /countries/{id}: Updates an existing Country by ID.
- DELETE /countries/{id}: Deletes an existing Country by ID.
- GET /covid-stats/{id}: Returns a CovidStatistic by ID.
- GET /covid-stats: Returns a list of CovidStatistics.
- POST /covid-stats: Creates a new CovidStatistic.
- PUT /covid-stats/{id}: Updates an existing CovidStatistic by ID.
- DELETE /covid-stats/{id}: Deletes an existing CovidStatistic by ID.
- GET /users/{userid}/monitored-countries: Returns a list of monitored countries for a User by ID.
- POST /users/{userid}/monitored-countries: Adds a new monitored country for a User by ID.
- DELETE /users/{userid}/monitored-countries/{countryId}: Removes a monitored country for a User by ID and Country ID.
- GET /countries/top-by-case-type/{caseType}/{limit}/{userId}: Returns a list of top countries by case type for a User by ID.
- GET /countries/{countryId}/death-percentage: Returns the death percentage for a Country by ID.
- POST /register: Registers a new user.
- POST /login: Logs in a user.
- DELETE /users/{userId}: Deletes a user by ID.
- PUT /users/{userId}: Updates a user by ID.
- POST /refresh-covid-data: Refreshes COVID data for all countries.

 * Addition/Updating a new country body looks like this:
 ```
{
    "name": "Canada",
    "code": "CA"
}
```
* Addition/Updating a single covid stat looks like this:
```
{
    "id": 1,
    "country_id": 1,
    "date": "2022-03-29T00:00:00Z",
    "confirmed_cases": 5000,
    "deaths": 200,
    "recovered": 3000
}  
```

* Addition/Updating monitored countries:
```
{
    "countryId": 1
}
```
* Registering a new user:

```
{
    "username": "examplename",
    "email": "examplename@example.com",
    "password": "Password1234!"
}
```

* Logging in:
```
{
   "username": "examplename",
   "password": "Password1234!"
}
```

