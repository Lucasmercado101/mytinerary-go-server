package endpoints

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"quickstart/database"
)

type CityJSON struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func City(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	switch r.Method {
	case "GET":
		id := r.URL.Path[len("/cities/"):]
		log.Printf("id: %s\n", id)

		row := database.Db.QueryRow("SELECT name, id FROM cities WHERE id = $1", id)
		var city CityJSON
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
		id := r.URL.Path[len("/cities/"):]
		log.Printf("id: %s\n", id)

		var city CityJSON
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&city); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		log.Printf("City: %+v", city)

		err := database.Db.QueryRow("UPDATE cities SET name = $1 WHERE id = $2 RETURNING *", city.Name, id).Scan(&city.Id, &city.Name)
		if err != nil {
			panic(err)
		}

		json.NewEncoder(w).Encode(city)

	case "DELETE":
		id := r.URL.Path[len("/cities/"):]
		log.Printf("id: %s\n", id)

		_, err := database.Db.Exec("DELETE FROM cities WHERE id = $1", id)
		if err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
