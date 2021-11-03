package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"quickstart/database"
	"quickstart/endpoints"

	"github.com/gorilla/mux"
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
	log.SetOutput(os.Stdout) // Set log output to standard output

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

	r := mux.NewRouter()

	s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/static/").Handler(s)

	r.HandleFunc("/cities", returnsJSONMiddleware(endpoints.Cities))
	r.HandleFunc("/cities/{cityId}", returnsJSONMiddleware(endpoints.City))

	// CORS
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#the_http_response_headers add the Vary header
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			// if Preflight
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%v", (60*5)))
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)

		})
	})

	r.HandleFunc("/auth/login", endpoints.Login)
	r.HandleFunc("/auth/register", endpoints.Register)
	r.HandleFunc("/auth/isLoggedIn", endpoints.IsLoggedIn)
	r.HandleFunc("/auth/logout", endpoints.Logout)

	r.HandleFunc("/itinerary", endpoints.Itineraries).Methods("POST")
	r.HandleFunc("/itinerary/{itineraryId}", returnsJSONMiddleware(endpoints.Itinerary)).Methods("GET", "PUT", "DELETE", "PATCH")

	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":8001", nil))

}

type endpoint func(http.ResponseWriter, *http.Request)

func returnsJSONMiddleware(fn endpoint) endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fn(w, r)
	}
}
