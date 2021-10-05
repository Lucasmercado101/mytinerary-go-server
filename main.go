package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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

var db *sql.DB

func dbInit() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	newDb, err := sql.Open("postgres", psqlInfo)
	db = newDb
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")

	db.Exec(`
	CREATE TABLE IF NOT EXISTS cities
	(
		id SERIAL NOT NULL PRIMARY KEY,
		name VARCHAR(40)
	)
	`)
}

func citiesEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT name, id FROM cities")
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		var cities []City
		for rows.Next() {
			var city City
			if err := rows.Scan(&city.Name, &city.Id); err != nil {
				panic(err)
			}
			cities = append(cities, city)
		}
		if err := rows.Err(); err != nil {
			panic(err)
		}

		// if cities is empty, return empty array
		if len(cities) == 0 {
			w.Write([]byte("[]"))
		} else {
			json.NewEncoder(w).Encode(cities)
		}

	case "POST":
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be 'application/json'", http.StatusUnsupportedMediaType)
			return
		}

		var city City
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&city); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		log.Printf("City: %+v", city)

		_, err := db.Exec("INSERT INTO cities (name) VALUES ($1)", city.Name)
		if err != nil {
			panic(err)
		}

		// w.WriteHeader(http.StatusCreated)
	}
}

func cityEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	switch r.Method {
	case "GET":
		id := r.URL.Path[len("/cities/"):]
		log.Printf("id: %s\n", id)

		row := db.QueryRow("SELECT name, id FROM cities WHERE id = $1", id)
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

		// case "PUT":
		// 	id := r.URL.Path[len("/cities/"):]
		// 	log.Printf("id: %s\n", id)

		// 	var city City
		// 	decoder := json.NewDecoder(r.Body)
		// 	if err := decoder.Decode(&city); err != nil {
		// 		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		// 		return
		// 	}
		// 	log.Printf("City: %+v", city)

		// 	_, err := db.Exec("UPDATE cities SET name = $1 WHERE id = $2", city.Name, id)
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	w.WriteHeader(http.StatusNoContent)

		// case "DELETE":
		// 	id := r.URL.Path[len("/cities/"):]
		// 	log.Printf("id: %s\n", id)

		// 	_, err := db.Exec("DELETE FROM cities WHERE id = $1", id)
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	w.WriteHeader(http.StatusNoContent)
	}
}

func main() {
	dbInit()
	defer db.Close()

	http.HandleFunc("/cities", citiesEndpoint)
	http.HandleFunc("/cities/", cityEndpoint)

	log.Fatal(http.ListenAndServe(":8001", nil))

}
