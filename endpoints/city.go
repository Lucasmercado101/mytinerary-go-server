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
	cityId := mux.Vars(r)["cityId"]

	switch r.Method {
	case "GET":
		// TODO: validation

		type ItineraryCommentJSON struct {
			Id           int    `json:"-"`
			Author_id    int    `json:"authorId"`
			Itinerary_id int    `json:"itineraryId"`
			Comment      string `json:"comment"`
		}

		type itinerary struct {
			Id         int                    `json:"id"`
			Title      string                 `json:"title"`
			Time       int                    `json:"time"`
			Price      int                    `json:"price"`
			Activities pq.StringArray         `json:"activities"`
			Hashtags   pq.StringArray         `json:"hashtags"`
			Comments   []ItineraryCommentJSON `json:"comments"`
			Creator    itineraryCreator       `json:"creator"`
			CityId     int                    `json:"-"`
		}

		rows, err := database.Db.Query(`
	SELECT id,
		title,
		time,
		price,
		activities,
		hashtags,
		user_id,
    	profile_pic,
		city_id
	FROM itinerary
		INNER JOIN (
			SELECT id as user_id,
				profile_pic
			FROM users
		) AS users ON itinerary.creator = users.user_id
	WHERE itinerary.city_id = $1`, cityId)

		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		var itineraries []itinerary

		for rows.Next() {

			var itinerary itinerary

			err := rows.Scan(&itinerary.Id, &itinerary.Title,
				&itinerary.Time, &itinerary.Price,
				&itinerary.Activities, &itinerary.Hashtags,
				&itinerary.Creator.User_id, &itinerary.Creator.Profile_pic, &itinerary.CityId)
			if err != nil {
				log.Fatal(err)
			}

			itineraries = append(itineraries, itinerary)
		}

		if len(itineraries) == 0 {
			w.Write([]byte("[]"))
			return
		}

		var itineraryIds []int
		for _, itinerary := range itineraries {
			itineraryIds = append(itineraryIds, itinerary.Id)
		}

		rows, err = database.Db.Query(`
		SELECT id,
			author_id,
			comment,
			itinerary_id
		FROM itinerary_comments
			INNER JOIN (
				SELECT id as ic_id, author_id, comment
				FROM itinerary_comment
			) AS ic ON ic.ic_id = itinerary_comments.comment_id
		WHERE itinerary_comments.itinerary_id = ANY($1::int[])
			`, pq.Array(itineraryIds))

		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		var comments []ItineraryCommentJSON

		for rows.Next() {

			var comment ItineraryCommentJSON

			err := rows.Scan(&comment.Id, &comment.Author_id, &comment.Comment, &comment.Itinerary_id)
			if err != nil {
				log.Fatal(err)
			}

			comments = append(comments, comment)
		}

		for index, itinerary := range itineraries {
			for _, comment := range comments {
				if comment.Itinerary_id == itinerary.Id {
					itineraries[index].Comments = append(itineraries[index].Comments, comment)
				}
			}
		}

		json.NewEncoder(w).Encode(itineraries)

	case "POST":

		session, err := database.IsUserLoggedIn(r)
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

		var itinerary itinerary
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&itinerary); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		_, err = database.Db.Exec(`
		INSERT INTO itinerary (title, time, price, activities, hashtags, creator, city_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, itinerary.Title, itinerary.Time, itinerary.Price,
			itinerary.Activities, itinerary.Hashtags, session.User_id, cityId)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
