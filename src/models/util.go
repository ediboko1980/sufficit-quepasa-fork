package models

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"github.com/go-chi/jwtauth"
)

// GetUser gets the user_id from the JWT and finds the
// corresponding user in the database
func GetUser(r *http.Request) (QPUser, error) {
	var user QPUser
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return user, err
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return user, errors.New("User ID missing")
	}

	return WhatsAppService.DB.User.FindByID(userID)
}

// CleanPhoneNumber removes all non-numeric characters from a string
func CleanPhoneNumber(number string) (string, error) {
	var out string
	if strings.HasSuffix(number, "@g.us") {
		return out, fmt.Errorf("this id is a group, cant be converted to phone number")
	}

	return GetPhoneByID(number)
}

// Usado também para identificar o número do bot
// Meramente visual
func GetPhoneByID(id string) (out string, err error) {
	spacesRemoved := strings.Replace(id, " ", "", -1)
	re, err := regexp.Compile(`\d*`)
	matches := re.FindAllString(spacesRemoved, -1)
	if len(matches) > 0 {
		out = matches[0]
	}
	return out, err
}