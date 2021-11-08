package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"quickstart/database"
	"quickstart/endpoints"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type City struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

var Db *sql.DB

// Cron job to delete expired sessions every 24 hours
func deleteOldSessions() {
	for range time.Tick(time.Hour * 24) {
		rows, err := Db.Query("SELECT session_id, id FROM sessions WHERE expiration < NOW()")

		if err != nil {
			panic(err)
		}

		for rows.Next() {
			var sessionId string
			var id int
			err = rows.Scan(&sessionId, &id)
			if err != nil {
				panic(err)
			}
			log.Println("Deleting session:", sessionId)
			_, err := Db.Exec("DELETE FROM sessions WHERE id = $1", id)
			if err != nil {
				panic(err)
			}
		}
	}

}

func main() {
	godotenv.Load()

	log.SetOutput(os.Stdout) // Set log output to standard output

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_DBNAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
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

	// CORS
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO Allow Origin is string array or origin
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Vary", "Origin")
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

	s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/static/").Handler(s)

	r.HandleFunc("/cities", returnsJSONMiddleware(endpoints.Cities))
	r.HandleFunc("/cities/{cityId:[0-9]+}", returnsJSONMiddleware(endpoints.City))
	r.HandleFunc("/cities/{cityId:[0-9]+}/itinerary", returnsJSONMiddleware(endpoints.CityItineraries))

	r.HandleFunc("/auth/login", returnsJSONMiddleware(endpoints.Login))
	r.HandleFunc("/auth/register", endpoints.Register)
	r.HandleFunc("/auth/isLoggedIn", endpoints.IsLoggedIn)
	r.HandleFunc("/auth/logout", endpoints.Logout)

	r.HandleFunc("/itinerary", endpoints.Itineraries).Methods("POST")
	r.HandleFunc("/itinerary/{itineraryId:[0-9]+}", returnsJSONMiddleware(endpoints.Itinerary)).Methods("GET", "PUT", "DELETE")
	r.HandleFunc("/itinerary/{itineraryId:[0-9]+}/comment", returnsJSONMiddleware(endpoints.ItineraryComment))

	http.Handle("/", r)

	// Cron job
	go deleteOldSessions()
	log.Fatal(http.ListenAndServe(":8001", nil))
}

type endpoint func(http.ResponseWriter, *http.Request)

func returnsJSONMiddleware(fn endpoint) endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fn(w, r)
	}
}
