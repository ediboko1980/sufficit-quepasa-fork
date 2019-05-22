package models

import (
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/go-chi/jwtauth"
	"github.com/jmoiron/sqlx"
)

// GetDB returns a database connection for connection the string
// set in the DB_CONNECTION environment variable
func GetDB() *sqlx.DB {
	connection := os.Getenv("DB_CONNECTION")
	db, err := sqlx.Connect("postgres", connection)

	if err != nil {
		log.Fatalln(err)
	}

	return db
}

// GetUser gets the user_id from the JWT and finds the
// corresponding user in the database
func GetUser(r *http.Request) (User, error) {
	var user User
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return user, err
	}

	userID := claims["user_id"].(string)
	if userID == "" {
		return user, errors.New("User ID missing")
	}

	return FindUserByID(GetDB(), userID)
}

// CleanPhoneNumber removes all non-numeric characters from a string
func CleanPhoneNumber(number string) string {
	var out string
	re := regexp.MustCompile("\\d*")
	matches := re.FindAllString(number, -1)
	if len(matches) > 0 {
		out = matches[0]
	}
	return out
}
