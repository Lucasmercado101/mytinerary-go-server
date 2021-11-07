package endpoints

import (
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"quickstart/database"
)

func Cities(w http.ResponseWriter, r *http.Request) {

	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	switch r.Method {
	case "GET":
		rows, err := database.Db.Query("SELECT * FROM city")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// if cities is empty, return empty array
		if len(cities) == 0 {
			w.Write([]byte("[]"))
		} else {
			json.NewEncoder(w).Encode(cities)
		}

	case "POST":

		_, err := database.IsUserLoggedIn(r)
		if err != nil {
			switch err {
			case database.ErrNoCookie:
				w.WriteHeader(http.StatusUnauthorized)
				return

			case database.ErrUnauthorized:
				w.WriteHeader(http.StatusUnauthorized)
				return

			default:
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var city CityJSON

		switch mediaType {
		case "application/json":
			if err := json.NewDecoder(r.Body).Decode(&city); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

		case "application/x-www-form-urlencoded":
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form", http.StatusBadRequest)
				return
			}
			city.Name = r.Form.Get("name")
			city.Country = r.Form.Get("country")

		case "multipart/form-data":
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				http.Error(w, "Invalid form", http.StatusBadRequest)
				return
			}
			city.Name = r.Form.Get("name")
			city.Country = r.Form.Get("country")
		}

		log.Printf("%+v\n", city)

		err = database.Db.QueryRow("INSERT INTO city (name, country) VALUES ($1, $2) RETURNING *", city.Name, city.Country).Scan(&city.Id, &city.Name, &city.Country)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		json.NewEncoder(w).Encode(city)
	}
}
