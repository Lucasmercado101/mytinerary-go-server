package endpoints

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"quickstart/database"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

type CityJSON struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
}

type itinerary struct {
	Id         int              `json:"id"`
	Title      string           `json:"title"`
	Time       int              `json:"time"`
	Price      int              `json:"price"`
	Activities pq.StringArray   `json:"activities"`
	Hashtags   pq.StringArray   `json:"hashtags"`
	Creator    itineraryCreator `json:"creator"`
}

type itineraryCreator struct {
	User_id     int    `json:"userId"`
	Profile_pic string `json:"profilePic"`
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

func CityItineraries(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
	id := mux.Vars(r)["cityId"]

	// TODO: validation

	rows, err := database.Db.Query(`
	SELECT id,
		title,
		time,
		price,
		activities,
		hashtags,
		user_id,
    	profile_pic
	FROM itinerary
		INNER JOIN (
			SELECT id as user_id,
				profile_pic
			FROM users
		) AS users ON itinerary.creator = users.user_id
	WHERE itinerary.id = $1`, id)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		rows.Close()
		return
	}

	defer rows.Close()

	var itineraries []itinerary

	for rows.Next() {

		var itinerary itinerary

		err := rows.Scan(&itinerary.Id, &itinerary.Title,
			&itinerary.Time, &itinerary.Price,
			&itinerary.Activities, &itinerary.Hashtags,
			&itinerary.Creator.User_id, &itinerary.Creator.Profile_pic)
		if err != nil {
			log.Fatal(err)
		}

		itineraries = append(itineraries, itinerary)
	}

	json.NewEncoder(w).Encode(itineraries)
}
