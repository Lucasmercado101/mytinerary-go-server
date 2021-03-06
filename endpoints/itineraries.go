package endpoints

import (
	"encoding/json"
	"log"
	"net/http"
	"quickstart/database"

	"github.com/lib/pq"
)

type itineraryInput struct {
	CityId     *int     `json:"cityId"`
	AuthorId   *int     `json:"authorId"`
	Title      string   `json:"title"`
	Duration   *int     `json:"duration"`
	Price      *int     `json:"price"`
	Tags       []string `json:"tags"`
	Activities []string `json:"activities"`
}

func Itineraries(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
	var input itineraryInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validation
	if input.CityId == nil {
		http.Error(w, "Missing cityId", http.StatusBadRequest)
		return
	}
	if input.AuthorId == nil {
		http.Error(w, "Missing authorId", http.StatusBadRequest)
		return
	}
	if input.Title == "" {
		http.Error(w, "Missing title", http.StatusBadRequest)
		return
	}
	if input.Duration == nil {
		http.Error(w, "Missing duration", http.StatusBadRequest)
		return
	}
	if input.Price == nil {
		http.Error(w, "Missing price", http.StatusBadRequest)
		return
	}
	if len(input.Tags) == 0 {
		http.Error(w, "Missing tags", http.StatusBadRequest)
		return
	}
	if len(input.Activities) == 0 {
		http.Error(w, "Missing activities", http.StatusBadRequest)
		return
	}

	if len(input.Tags) > 3 {
		http.Error(w, "Too many tags", http.StatusBadRequest)
		return
	}

	_, err = database.Db.Exec(`
	INSERT INTO itinerary (
        title,
        creator,
        time,
        price,
        activities,
        hashtags,
        city_id
    )
	VALUES ($1, $2, $3, $4, $5, $6, $7)
`,
		input.Title,
		input.AuthorId,
		input.Duration,
		input.Price,
		pq.Array(input.Activities),
		pq.Array(input.Tags),
		input.CityId)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
