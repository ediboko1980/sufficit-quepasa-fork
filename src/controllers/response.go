package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sufficit/sufficit-quepasa-fork/models"
)

type errorResponse struct {
	Result string `json:"result"`
}

func respondSuccess(w http.ResponseWriter, res interface{}) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func respondBadRequest(w http.ResponseWriter, err error) {
	log.Println("!Request Bad Format: ", err)

	respondError(w, err, http.StatusBadRequest)
}

func respondUnauthorized(w http.ResponseWriter, err error) {
	log.Println("!Request Unauthorized: ", err)

	respondError(w, err, http.StatusUnauthorized)
}

func respondNotFound(w http.ResponseWriter, err error) {
	log.Println("!Request Not found: ", err)

	respondError(w, err, http.StatusNotFound)
}

/// Usado para avisar que o bot ainda não esta pronto
func respondNotReady(w http.ResponseWriter, err error) {
	respondError(w, err, http.StatusServiceUnavailable)
}

func respondServerError(bot models.QPBot, w http.ResponseWriter, err error) {
	log.Printf("(%s) !Request Server error: %s", bot.GetNumber(), err)
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
