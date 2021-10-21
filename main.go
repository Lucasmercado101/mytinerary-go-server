package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"quickstart/database"
	cities "quickstart/endpoints"

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

func cityEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	switch r.Method {
	case "GET":
		id := r.URL.Path[len("/cities/"):]
		log.Printf("id: %s\n", id)

		row := Db.QueryRow("SELECT name, id FROM cities WHERE id = $1", id)
		var city City
		err := row.Scan(&city.Name, &city.Id)
		if err == sql.ErrNoRows {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		if err != nil {
			panic(err)
		}

		json.NewEncoder(w).Encode(city)

	case "PUT":
		// full update
		id := r.URL.Path[len("/cities/"):]
		log.Printf("id: %s\n", id)

		var city City
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&city); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		log.Printf("City: %+v", city)

		_, err := Db.Exec("UPDATE cities SET name = $1 WHERE id = $2", city.Name, id)
		if err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)

		// case "PATCH":
		// partial update
		// case "DELETE":
		// 	id := r.URL.Path[len("/cities/"):]
		// 	log.Printf("id: %s\n", id)

		// 	_, err := Db.Exec("DELETE FROM cities WHERE id = $1", id)
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	w.WriteHeader(http.StatusNoContent)
	}
}

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

	err = Db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	defer Db.Close()

	http.HandleFunc("/cities", returnsJSONMiddleware(cities.CitiesEndpoint))
	http.HandleFunc("/cities/", returnsJSONMiddleware(cityEndpoint))

	log.Fatal(http.ListenAndServe(":8001", nil))

}

type endpoint func(http.ResponseWriter, *http.Request)

func returnsJSONMiddleware(fn endpoint) endpoint {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fn(w, r)
	}
}
