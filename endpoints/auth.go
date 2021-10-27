package endpoints

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"quickstart/database"

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

// func Login(w http.ResponseWriter, r *http.Request) {
// 	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)

// 	var creds AuthCreds
// 	err := json.NewDecoder(r.Body).Decode(&creds)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("%v", creds)
// 	hash, err := hashPassword(creds.Password)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("%v", hash)
// 	if checkPasswordHash(creds.Password, hash) {
// 		fmt.Println("Password is correct")
// 	} else {
// 		fmt.Println("Password is incorrect")
// 	}
// }

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
