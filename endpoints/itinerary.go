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

	// TODO: validation, if city & creator exist

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

type itineraryCommentInput struct {
	Content  string `json:"content"`
	AuthorId int    `json:"authorId"`
}

func ItineraryComment(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	itineraryId := mux.Vars(r)["itineraryId"]

	var input itineraryCommentInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validation
	if input.Content == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// TODO: more validation

	tx, err := database.Db.Begin()

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var newCommentId int

	err = tx.QueryRow(`
	INSERT INTO itinerary_comment (author_id, comment)
	VALUES ($1, $2)
	RETURNING id
	`, input.AuthorId, input.Content).Scan(&newCommentId)

	if err != nil {
		tx.Rollback()
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err != nil {
		tx.Rollback()
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`
	INSERT INTO itinerary_comments (itinerary_id, comment_id)
	VALUES ($1, $2)
	`, itineraryId, newCommentId)

	if err != nil {
		tx.Rollback()
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx.Commit()

}
