package endpoints

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"quickstart/database"

	"github.com/gorilla/mux"
)

type CityJSON struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
}

func City(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	id := mux.Vars(r)["cityId"]
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	switch r.Method {
	case "GET":

		row := database.Db.QueryRow("SELECT * FROM city WHERE id = $1", id)
		var city CityJSON
		err := row.Scan(&city.Id, &city.Name, &city.Country)
		if err == sql.ErrNoRows {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		if err != nil {
			panic(err)
		}

		json.NewEncoder(w).Encode(city)

	case "PUT":

		var city CityJSON
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&city); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		log.Printf("City: %+v", city)

		err := database.Db.QueryRow(`
		UPDATE city
		SET name = $1, country = $2
		WHERE id = $3
		RETURNING *
		`, city.Name, city.Country, id).Scan(&city.Id, &city.Name, &city.Country)

		if err != nil {
			panic(err)
		}

		json.NewEncoder(w).Encode(city)

	case "DELETE":

		_, err := database.Db.Exec("DELETE FROM city WHERE id = $1", id)
		if err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
