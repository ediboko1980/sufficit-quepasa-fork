package models

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

// GetDB returns a database connection for the given
// database environment variables
func GetDB() *sqlx.DB {
	host := os.Getenv("PGHOST")
	database := os.Getenv("PGDATABASE")
	port := os.Getenv("PGPORT")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	ssl := os.Getenv("PGSSLMODE")
	connection := fmt.Sprintf("host=%s dbname=%s port=%s user=%s password=%s sslmode=%s",
		host, database, port, user, password, ssl)
	db, err := sqlx.Connect("postgres", connection)
	db.DB.SetMaxIdleConns(500)
	db.DB.SetMaxOpenConns(1000)
	db.DB.SetConnMaxLifetime(30 * time.Second)

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

	userID, ok := claims["user_id"].(string)
	if !ok {
		return user, errors.New("User ID missing")
	}

	return FindUserByID(GetDB(), userID)
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

// MigrateToLatest updates the database to the latest schema
func MigrateToLatest() error {
	log.Println("Migrating database (if necessary)")

	workDir, err := os.Getwd()
	if err != nil {
		log.Printf("Migrating database error on working dir: %s", workDir)
		return err
	}

	var fullPath string
	if runtime.GOOS == "windows" {
		log.Println("Migrating database on Windows")

		// windows ===================
		leadingWindowsUnit, _ := filepath.Rel("z:\\", workDir)
		migrationsDir := filepath.Join(leadingWindowsUnit, "migrations")
		fullPath = fmt.Sprintf("file:///%s", strings.ReplaceAll(migrationsDir, "\\", "/"))
	} else {
		// linux ===================
		migrationsDir := filepath.Join(workDir, "migrations")
		fullPath = fmt.Sprintf("file://%s", migrationsDir)
	}

	host := os.Getenv("PGHOST")
	database := os.Getenv("PGDATABASE")
	port := os.Getenv("PGPORT")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	ssl := os.Getenv("PGSSLMODE")
	connection := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, database, ssl)

	m, err := migrate.New(fullPath, connection)
	if err != nil {
		log.Printf("Migrating database error on: %s", fullPath)
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Printf("Migrating database error on: %s", fullPath)
		return err
	}

	return nil
}
