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

func Itinerary(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
	itineraryId := mux.Vars(r)["itineraryId"]

	switch r.Method {
	case "GET":
		var itinerary struct {
			Id         int            `json:"id"`
			Title      string         `json:"title"`
			Creator    int            `json:"creator"`
			Time       string         `json:"time"`
			Price      string         `json:"price"`
			Activities pq.StringArray `json:"activities"`
			Hashtags   pq.StringArray `json:"hashtags"`
			CityId     int            `json:"cityId"`
			Comments   []struct {
				Id           int    `json:"id"`
				Itinerary_id int    `json:"-"`
				Comment      string `json:"comment"`
				Author       struct {
					Id         int    `json:"id"`
					ProfilePic string `json:"profilePic"`
				} `json:"author"`
			} `json:"comments"`
		}

		err := database.Db.QueryRow("SELECT * FROM itinerary WHERE id = $1", itineraryId).Scan(
			&itinerary.Id,
			&itinerary.Title,
			&itinerary.Creator,
			&itinerary.Time,
			&itinerary.Price,
			&itinerary.Activities,
			&itinerary.Hashtags,
			&itinerary.CityId,
		)

		if err != nil {
			log.Println(err)
			// if city doesn't exist
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rows, err := database.Db.Query(`
		SELECT id,
			author_id,
			comment,
			itinerary_id,
			user_id,
			profile_pic
		FROM itinerary_comments
			INNER JOIN (
				SELECT id as ic_id,
					author_id,
					comment
				FROM itinerary_comment
			) AS ic ON ic.ic_id = itinerary_comments.comment_id
			INNER JOIN (
				SELECT id as user_id,
					profile_pic
				FROM users
			) AS users ON users.user_id = ic.author_id
		WHERE itinerary_comments.itinerary_id = $1
		`, itineraryId)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer rows.Close()

		for rows.Next() {
			var comment struct {
				Id           int    `json:"id"`
				Itinerary_id int    `json:"-"`
				Comment      string `json:"comment"`
				Author       struct {
					Id         int    `json:"id"`
					ProfilePic string `json:"profilePic"`
				} `json:"author"`
			}

			err := rows.Scan(
				&comment.Id,
				&comment.Author.Id,
				&comment.Comment,
				&comment.Itinerary_id,
				&comment.Author.Id,
				&comment.Author.ProfilePic,
			)

			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			itinerary.Comments = append(itinerary.Comments, comment)
		}

		json.NewEncoder(w).Encode(itinerary)

	case "PUT":

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

		var itinerary struct {
			Title      string         `json:"title"`
			Time       int            `json:"time"`
			Price      int            `json:"price"`
			Activities pq.StringArray `json:"activities"`
			Hashtags   pq.StringArray `json:"hashtags"`
		}

		err = json.NewDecoder(r.Body).Decode(&itinerary)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = database.Db.Exec(`
		UPDATE itinerary
		SET title = $1,
			time = $2,
			price = $3,
			activities = $4,
			hashtags = $5
		WHERE id = $6
		`,
			itinerary.Title,
			itinerary.Time,
			itinerary.Price,
			itinerary.Activities,
			itinerary.Hashtags,
			itineraryId)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	case "DELETE":

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

		// Check if itinerary being deleted belongs to the user
		var creator int
		err = database.Db.QueryRow(`
		SELECT creator
		FROM itinerary
		WHERE id = $1
		`, itineraryId).Scan(&creator)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if creator != session.User_id {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// It belongs to the user, delete it

		_, err = database.Db.Exec(`
		DELETE FROM itinerary
		WHERE id = $1
		`, itineraryId)

		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type itineraryCommentInput struct {
	Content  string `json:"content"`
	AuthorId int    `json:"authorId"`
}

type itineraryCommentResponse struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
	Creator struct {
		CreatorId  int    `json:"creatorId"`
		ProfilePic string `json:"profilePic"`
	} `json:"creator"`
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

	err = tx.Commit()

	if err != nil {
		tx.Rollback()
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var comment itineraryCommentResponse

	err = database.Db.QueryRow(`
	SELECT id,
		comment,
		user_id,
		profile_pic
	FROM itinerary_comment
		INNER JOIN (
			SELECT id AS user_id,
				profile_pic
			FROM users
		) AS users ON users.user_id = itinerary_comment.author_id
	WHERE itinerary_comment.id = $1
	`, newCommentId).Scan(
		&comment.Id,
		&comment.Content,
		&comment.Creator.CreatorId,
		&comment.Creator.ProfilePic,
	)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comment)
}
