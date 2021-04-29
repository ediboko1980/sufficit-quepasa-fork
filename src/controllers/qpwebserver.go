package controllers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
)

func QPWebServerStart() {
	r := newRouter()
	webAPIPort := os.Getenv("WEBAPIPORT")
	if len(webAPIPort) == 0 {
		webAPIPort = "31000"
	}

	log.Printf("Starting Web Server on Port: %s", webAPIPort)
	http.ListenAndServe(":"+webAPIPort, r)
}

func newRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	//r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// web routes
	addWebRoutes(r)

	// api routes
	addAPIRoutes(r)

	// static files
	workDir, _ := os.Getwd()
	assetsDir := filepath.Join(workDir, "assets")
	fileServer(r, "/assets", http.Dir(assetsDir))

	return r
}

func addWebRoutes(r chi.Router) {
	tokenAuth := jwtauth.New("HS256", []byte(os.Getenv("SIGNING_SECRET")), nil)

	// authenticated web routes
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(authenticator)

		r.Get("/account", AccountFormHandler)
		r.Get("/bot/register", RegisterFormHandler)
		r.Post("/bot/register", RegisterHandler)
		r.Get("/bot/{botID}/verify/ws", VerifyHandler)
		r.Get("/bot/{botID}/verify", VerifyFormHandler)
		r.Post("/bot/delete", DeleteHandler)
		r.Post("/bot/cycle", CycleHandler)
		r.Get("/bot/{botID}", SendFormHandler)
		r.Get("/bot/{botID}/send", SendFormHandler)
		r.Post("/bot/{botID}/send", SendHandler)
		r.Get("/bot/{botID}/receive", ReceiveFormHandler)
	})

	// unauthenticated web routes
	r.Group(func(r chi.Router) {
		r.Get("/", IndexHandler)
		r.Get("/login", LoginFormHandler)
		r.Post("/login", LoginHandler)
		r.Get("/setup", SetupFormHandler)
		r.Post("/setup", SetupHandler)
		r.Get("/logout", LogoutHandler)
	})
}

func addAPIRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Get("/v1/bot/{token}", InfoAPIHandler)
		r.Post("/v1/bot/{token}/send", SendAPIHandler)
		r.Get("/v1/bot/{token}/receive", ReceiveAPIHandler)

		r.Post("/v1/bot/{token}/attachment", AttachmentHandler)
		r.Post("/v1/bot/{token}/webhook", WebHookHandler)
	})
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"
	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

func authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if token == nil || !token.Valid {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}