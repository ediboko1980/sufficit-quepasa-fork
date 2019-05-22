package controllers

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorResponse struct {
	Result string `json:"result"`
}

// RespondSuccess generates HTTP status code 200
func RespondSuccess(w http.ResponseWriter, res interface{}) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// RespondBadRequest generates HTTP status code 400
func RespondBadRequest(w http.ResponseWriter, err error) {
	log.Println("Bad request: ", err)

	respondError(w, err, http.StatusBadRequest)
}

// RespondUnauthorized generates HTTP status code 401
func RespondUnauthorized(w http.ResponseWriter, err error) {
	log.Println("Unauthorized: ", err)

	respondError(w, err, http.StatusUnauthorized)
}

// RespondNotFound generates HTTP status code 404
func RespondNotFound(w http.ResponseWriter, err error) {
	log.Println("Not found: ", err)

	respondError(w, err, http.StatusNotFound)
}

// RespondServerError generates HTTP status code 500
func RespondServerError(w http.ResponseWriter, err error) {
	log.Println("Server error: ", err)

	respondError(w, err, http.StatusInternalServerError)
}

func respondError(w http.ResponseWriter, err error, code int) {
	res := &errorResponse{
		Result: err.Error(),
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
