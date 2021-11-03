package endpoints

import (
	"encoding/json"
	"log"
	"net/http"
	"quickstart/database"

	"github.com/gorilla/mux"
)

func Itinerary(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	itineraryId := mux.Vars(r)["itineraryId"]

	var itinerary database.Itinerary

	database.Db.QueryRow("SELECT * FROM itinerary WHERE id = $1", itineraryId).Scan(
		&itinerary.Id,
		&itinerary.Title,
		&itinerary.Creator,
		&itinerary.Time,
		&itinerary.Price,
		&itinerary.Activities,
		&itinerary.Hashtags,
		&itinerary.CityId,
	)

	json.NewEncoder(w).Encode(itinerary)
}
