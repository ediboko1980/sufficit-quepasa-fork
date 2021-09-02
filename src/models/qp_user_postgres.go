package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type QPUserPostgres struct {
	db *sqlx.DB
}

func (source QPUserPostgres) Count() (int, error) {
	var count int
	err := source.db.Get(&count, "SELECT count(*) FROM users")
	return count, err
}

func (source QPUserPostgres) FindByID(ID string) (QPUser, error) {
	var user QPUser
	err := source.db.Get(&user, "SELECT * FROM users WHERE id = $1", ID)
	return user, err
}

func (source QPUserPostgres) FindByEmail(email string) (QPUser, error) {
	var user QPUser
	err := source.db.Get(&user, "SELECT * FROM users WHERE email = $1", email)
	return user, err
}

func (source QPUserPostgres) Exists(email string) (bool, error) {
	var count int
	err := source.db.Get(&count, "SELECT count(*) FROM users WHERE email = $1", email)
	return count > 0, err
}

func (source QPUserPostgres) Check(email string, password string) (QPUser, error) {
	var user QPUser
	var out QPUser
	err := source.db.Get(&user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		return out, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return out, err
	}

	out = user
	return out, err
}

func (source QPUserPostgres) Create(email string, password string) (QPUser, error) {
	var user QPUser
	userID := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return user, err
	}

	query := `INSERT INTO users
    (id, email, username, password, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := source.db.Exec(query, userID, email, email, string(hashed), now, now); err != nil {
		return user, err
	}

	return source.FindByID(userID)
}
