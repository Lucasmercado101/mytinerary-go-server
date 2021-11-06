package endpoints

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"quickstart/database"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthCreds struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const cookieMaxAge = time.Hour * 24

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
		id          string
		password    string
		profile_pic sql.NullString
	}
	err = database.Db.QueryRow("SELECT id, password, profile_pic FROM users WHERE username = $1", creds.Username).Scan(&dbUser.id, &dbUser.password, &dbUser.profile_pic)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		panic(err)
	}

	var expirationTime = time.Now().Add(cookieMaxAge)

	if checkPasswordHash(creds.Password, dbUser.password) {
		cookie := http.Cookie{
			Name:    "sid",
			Value:   uuid.New().String(),
			Path:    "/",
			Expires: expirationTime,
			MaxAge:  int(cookieMaxAge.Seconds()),
		}
		http.SetCookie(w, &cookie)
		_, err := database.Db.Exec("INSERT INTO sessions (user_id, session_id, expiration) VALUES ($1, $2, $3)", dbUser.id, cookie.Value, expirationTime)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var userDTO struct {
			Id          int    `json:"id"`
			Username    string `json:"username"`
			Profile_pic string `json:"profilePic"`
		}
		userDTO.Username = creds.Username
		userDTO.Profile_pic = dbUser.profile_pic.String
		if dbUser.profile_pic.Valid {
			userDTO.Profile_pic = dbUser.profile_pic.String
		} else {
			userDTO.Profile_pic = ""
		}

		json.NewEncoder(w).Encode(userDTO)

	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func Register(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
	// TODO: validation / error handling

	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	if mediaType == "multipart/form-data" {
		w.Header().Set("Content-Type", "application/json")

		// max size 25MB
		r.ParseMultipartForm(25 * 1024 * 1024)

		username := r.FormValue("username")

		// check if username is already taken
		var dbUser struct {
			id string
		}
		err = database.Db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&dbUser.id)
		if err != nil {
			if err == sql.ErrNoRows {

				password := r.FormValue("password")

				pfpFile, header, err := r.FormFile("profilePic")
				if err != nil {
					log.Fatalln(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				defer pfpFile.Close()

				cwd, err := os.Getwd()
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				imagesDir := filepath.Join(cwd, "static", "images")

				// create folder if it doesn't exist
				if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
					err = os.MkdirAll(imagesDir, 0755)
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
					}
				}

				var imageFileName = uuid.New().String() + filepath.Ext(header.Filename)

				storedImagePath := filepath.Join(imagesDir, imageFileName)

				f, err := os.Create(storedImagePath)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				defer f.Close()

				_, err = io.Copy(f, pfpFile)

				if err != nil {
					// delete file
					os.Remove(storedImagePath)
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// begin storing in DB
				hash, err := hashPassword(password)
				if err != nil {
					// delete file
					os.Remove(storedImagePath)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				var scheme string
				if r.TLS != nil { // https://github.com/golang/go/issues/28940#issuecomment-441749380
					scheme = "https"
				} else {
					scheme = "http"
				}

				var imageUrl = scheme + "://" + r.Host + "/static/images/" + imageFileName

				_, err = database.Db.Exec("INSERT INTO users (username, password, profile_pic) VALUES ($1, $2, $3)", username, hash, imageUrl)

				if err != nil {
					// delete file
					os.Remove(storedImagePath)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				json.NewEncoder(w).Encode(struct {
					Username   string `json:"username"`
					ProfilePic string `json:"profilePic"`
				}{
					Username:   username,
					ProfilePic: storedImagePath,
				})

			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusConflict)
		}
	} else {

		var creds AuthCreds
		err = json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			panic(err)
		}

		err = database.Db.QueryRow("SELECT id FROM users WHERE username = $1", creds.Username).Scan(&AuthCreds{})
		if err != nil {
			if err == sql.ErrNoRows { // user doesn't exist so we can create it
				hash, err := hashPassword(creds.Password)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				_, err = database.Db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", creds.Username, hash)

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else { // user exists
				w.WriteHeader(http.StatusForbidden) // https://stackoverflow.com/a/34458500
			}
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
