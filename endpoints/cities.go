package endpoints

import (
	"encoding/json"
	"log"
	"net/http"
	"quickstart/database"
)

type City struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func Cities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	switch r.Method {
	case "GET":
		rows, err := database.Db.Query("SELECT name, id FROM cities")
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
		if err := json.NewDecoder(r.Body).Decode(&city); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		log.Printf("City: %+v", city)

		err := database.Db.QueryRow("INSERT INTO cities (name) VALUES ($1) RETURNING *", city.Name).Scan(&city.Id, &city.Name)
		if err != nil {
			panic(err)
		}

		json.NewEncoder(w).Encode(city)
	}
}
