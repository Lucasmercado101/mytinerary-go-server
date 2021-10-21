package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"quickstart/database"
	"quickstart/endpoints"

	_ "github.com/lib/pq"
)

type City struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "12345678"
	dbname   = "go_test"
)

var Db *sql.DB

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	newDb, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	Db = newDb
	database.Db = newDb

	err = newDb.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	defer newDb.Close()

	http.HandleFunc("/cities", returnsJSONMiddleware(endpoints.Cities))
	http.HandleFunc("/cities/", returnsJSONMiddleware(endpoints.City))

	log.Fatal(http.ListenAndServe(":8001", nil))

}

type endpoint func(http.ResponseWriter, *http.Request)

func returnsJSONMiddleware(fn endpoint) endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fn(w, r)
	}
}
