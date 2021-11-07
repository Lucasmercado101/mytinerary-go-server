package endpoints

import (
	"encoding/json"
	"log"
	"net/http"
	"quickstart/database"
)

func Cities(w http.ResponseWriter, r *http.Request) {

	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	switch r.Method {
	case "GET":
		rows, err := database.Db.Query("SELECT * FROM city")
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		var cities []CityJSON
		for rows.Next() {
			var city CityJSON
			if err := rows.Scan(&city.Id, &city.Name, &city.Country); err != nil {
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

		var city CityJSON
		if err := json.NewDecoder(r.Body).Decode(&city); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		log.Printf("City: %+v", city)

		err := database.Db.QueryRow("INSERT INTO city (name, country) VALUES ($1, $2) RETURNING *", city.Name, city.Country).Scan(&city.Id, &city.Name, &city.Country)
		if err != nil {
			panic(err)
		}

		json.NewEncoder(w).Encode(city)
	}
}
