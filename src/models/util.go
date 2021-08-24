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
	"strconv"
	"strings"

	"github.com/go-chi/jwtauth"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

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
func MigrateToLatest() (err error) {
	strMigrations := os.Getenv("MIGRATIONS")
	if len(strMigrations) == 0 {
		return
	}

	var fullPath string
	boolMigrations, err := strconv.ParseBool(strMigrations)
	if err == nil {
		// Caso false, migrações não habilitadas
		// Retorna sem problemas
		if !boolMigrations {
			return
		}
	} else {
		fullPath = strMigrations
	}

	log.Println("Migrating database (if necessary)")
	if boolMigrations {
		workDir, err := os.Getwd()
		if err != nil {
			return err
		}

		if runtime.GOOS == "windows" {
			log.Println("Migrating database on Windows")

			// windows ===================
			leadingWindowsUnit, _ := filepath.Rel("z:\\", workDir)
			migrationsDir := filepath.Join(leadingWindowsUnit, "migrations")
			fullPath = fmt.Sprintf("file:///%s", strings.ReplaceAll(migrationsDir, "\\", "/"))
		} else {
			// linux ===================
			migrationsDir := filepath.Join(workDir, "migrations")
			fullPath = fmt.Sprintf("file://%s", strings.TrimLeft(migrationsDir, "/"))
		}
	}

	config := GetDBConfig()
	connection := fmt.Sprintf("%s://%s:%s@tcp(%s:%s)/%s",
		config.Driver, config.User, config.Password, config.Host, config.Port, config.DataBase)

	m, err := migrate.New(fullPath, connection)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
