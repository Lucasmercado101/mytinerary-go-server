package endpoints

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"quickstart/database"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthCreds struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	var creds AuthCreds
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		panic(err)
	}

	var dbUser struct {
		id       string
		password string
	}
	err = database.Db.QueryRow("SELECT id, password FROM users WHERE username = $1", creds.Username).Scan(&dbUser.id, &dbUser.password)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		panic(err)
	}

	if checkPasswordHash(creds.Password, dbUser.password) {
		cookie := http.Cookie{
			Name:  "sid",
			Value: uuid.New().String(),
			Path:  "/",
		}
		http.SetCookie(w, &cookie)
		_, err := database.Db.Exec("INSERT INTO sessions (user_id, session_id, expiration) VALUES ($1, $2, $3)", dbUser.id, cookie.Value, time.Now().Add(time.Minute*2))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func Register(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

	var creds AuthCreds
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		panic(err)
	}

	err = database.Db.QueryRow("SELECT id FROM users WHERE username = $1", creds.Username).Scan(&AuthCreds{})
	if err != nil {
		if err == sql.ErrNoRows { // user doesn't exist so we can create it
			hash, err := hashPassword(creds.Password)
			if err != nil {
				panic(err)
			}
			_, err = database.Db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", creds.Username, hash)

			if err != nil {
				panic(err)
			}
		} else { // user exists
			w.WriteHeader(http.StatusForbidden) // https://stackoverflow.com/a/34458500
		}
	}
}

func IsLoggedIn(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		if err == http.ErrNoCookie {
			log.Printf("no cookie")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sessionId := cookie.Value
	var session struct {
		Id         string
		user_id    string
		session_id string
		expiration time.Time
	}
	err = database.Db.QueryRow("SELECT * FROM sessions WHERE session_id = $1", sessionId).Scan(&session.Id, &session.user_id, &session.session_id, &session.expiration)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("no session in db")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Println(err)
		return
	}

	if session.expiration.Before(time.Now()) {
		log.Printf("session expired, deleting session from db")
		w.WriteHeader(http.StatusUnauthorized)
		//delete the session if it exists in the db
		_, err := database.Db.Exec("DELETE FROM sessions WHERE session_id = $1", sessionId)
		if err != nil {
			log.Println(err)
		}
		return
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		if err == http.ErrNoCookie {
			log.Printf("no cookie")
			w.WriteHeader(http.StatusOK)
			return
		}
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sessionId := cookie.Value
	_, err = database.Db.Exec("DELETE FROM sessions WHERE session_id = $1", sessionId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}
