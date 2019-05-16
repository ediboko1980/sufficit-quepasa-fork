package common

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/jwtauth"
	"github.com/jmoiron/sqlx"
	"gitlab.com/digiresilience/link/quepasa/models"
)

// ParseJSONBody parses an HTTP request body into a map
func ParseJSONBody(r *http.Request) (map[string]interface{}, error) {
	var postParams map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&postParams); err != nil {
		return postParams, err
	}

	return postParams, nil
}

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

// RedirectToLogin generates HTTP status code 302 to "/login"
func RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}

// GetUser gets the user_id from the JWT and finds the
// corresponding user in the database
func GetUser(r *http.Request) (models.User, error) {
	var user models.User
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return user, err
	}

	userID := claims["user_id"].(string)
	if userID == "" {
		return user, errors.New("User ID missing")
	}

	return models.FindUserByID(GetDB(), userID)
}

// WebSocketProtocol determines which protocal to use based on
// the APP_ENV environment variable
func WebSocketProtocol() string {
	protocol := "wss"
	if os.Getenv("APP_ENV") == "development" {
		protocol = "ws"
	}
	return protocol
}