package controllers

import (
	"errors"
	"html/template"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
	"gitlab.com/digiresilience/link/quepasa/common"
	"gitlab.com/digiresilience/link/quepasa/models"
)

// Account home
type accountFormData struct {
	PageTitle string
	Bots      []models.Bot
	User      models.User
}

func AccountFormHandler(w http.ResponseWriter, r *http.Request) {
	user, err := common.GetUser(r)
	if err != nil {
		common.RedirectToLogin(w, r)
	}

	bots, err := models.FindAllBotsForUser(common.GetDB(), user.ID)
	if err != nil {
	}

	data := accountFormData{
		PageTitle: "Account",
		Bots:      bots,
		User:      user,
	}

	templates := template.Must(template.ParseFiles("views/layouts/main.tmpl", "views/account.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}

// Login form
type loginFormData struct {
	PageTitle string
}

func LoginFormHandler(w http.ResponseWriter, r *http.Request) {
	data := loginFormData{
		PageTitle: "Login",
	}

	templates := template.Must(template.ParseFiles("views/layouts/main.tmpl", "views/login.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	if email == "" || password == "" {
		common.RespondUnauthorized(w, errors.New("Missing username or password"))
		return
	}

	user, err := models.CheckUser(common.GetDB(), email, password)
	if err != nil {
		common.RespondUnauthorized(w, errors.New("Incorrect username or password"))
		return
	}

	tokenAuth := jwtauth.New("HS256", []byte(os.Getenv("SIGNING_SECRET")), nil)
	claims := jwt.MapClaims{"user_id": user.ID}
	jwtauth.SetIssuedNow(claims)
	jwtauth.SetExpiryIn(claims, 24*time.Hour)
	_, tokenString, _ := tokenAuth.Encode(claims)
	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		MaxAge:   60 * 60 * 24,
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/account", http.StatusFound)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    "",
		MaxAge:   0,
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)

	common.RedirectToLogin(w, r)
}

// Setup form
type passwordSuggestion struct {
	Suggestion string
}

type setupFormData struct {
	PageTitle             string
	ErrorMessage          string
	Email                 string
	EmailError            bool
	UserExistsError       bool
	EmailInvalidError     bool
	PasswordMatchError    bool
	PasswordStrengthError bool
	PasswordCrackTime     string
	PasswordSuggestions   []passwordSuggestion
}

func SetupFormHandler(w http.ResponseWriter, r *http.Request) {
	data := setupFormData{
		PageTitle:    "Setup",
		ErrorMessage: "Not good",
		PasswordSuggestions: []passwordSuggestion{
			passwordSuggestion{
				Suggestion: "Get a better password.",
			},
		},
	}

	templates := template.Must(template.ParseFiles("views/layouts/main.tmpl", "views/setup.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}

func SetupHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	if email == "" || password == "" {
		common.RespondBadRequest(w, errors.New("Email and password are required"))
		return
	}

	exists, err := models.CheckUserExists(common.GetDB(), email)
	if err != nil {
		common.RespondServerError(w, err)
		return
	}

	if exists {
		common.RespondServerError(w, errors.New("User exists"))
		return
	}

	_, err = models.CreateUser(common.GetDB(), email, password)
	if err != nil {
		common.RespondServerError(w, errors.New("Server error"))
		return
	}

	common.RedirectToLogin(w, r)
}
