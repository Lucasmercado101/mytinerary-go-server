package database

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/lib/pq"
)

var Db *sql.DB

type Itinerary struct {
	Id         int
	Title      string
	Creator    int
	Time       string
	Price      string
	Activities pq.StringArray
	Hashtags   pq.StringArray
	CityId     int
}

type session struct {
	Id         string
	User_id    string
	Session_id string
	Expiration time.Time
}

var ErrNoCookie = http.ErrNoCookie
var ErrInternalError = errors.New("internal error")
var ErrUnauthorized = errors.New("unauthorized")

func IsUserLoggedIn(r *http.Request) (session session, err error) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		if err == http.ErrNoCookie {
			log.Printf("no cookie")
			err = http.ErrNoCookie
			return
		}
		log.Println(err)
		err = ErrInternalError
		return
	}

	sessionId := cookie.Value
	err = Db.QueryRow("SELECT * FROM sessions WHERE session_id = $1", sessionId).Scan(&session.Id, &session.User_id, &session.Session_id, &session.Expiration)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("no session in db")
			err = ErrUnauthorized
			return
		}
		log.Println(err)
		return
	}

	if session.Expiration.Before(time.Now()) {
		log.Printf("session expired, deleting session from db")
		err = ErrUnauthorized
		//delete the session if it exists in the db
		_, dbErr := Db.Exec("DELETE FROM sessions WHERE session_id = $1", sessionId)
		if dbErr != nil {
			log.Println(err)
		}
		return
	}
	return
}
