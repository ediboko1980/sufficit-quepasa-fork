package controllers

import (
	"encoding/json"
	"net/http"
	"os"
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

// RedirectToLogin generates HTTP status code 302 to "/login"
func RedirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
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
