package models

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

type QPDataBase struct {
	Connection *sqlx.DB
	Sync       *sync.Mutex // Objeto de sinaleiro para evitar chamadas simultâneas a este objeto
}

var DBQuePasa *QPDataBase

func QPDataBaseInit() {
	sync := &sync.Mutex{}
	DBQuePasa = &QPDataBase{GetDB(), sync}
}

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

	// Tenta realizar a conexão
	if err != nil {
		log.Fatalln(err)
	}

	db.DB.SetMaxIdleConns(500)
	db.DB.SetMaxOpenConns(1000)
	db.DB.SetConnMaxLifetime(30 * time.Second)

	if err != nil {
		log.Fatalln(err)
	}

	return db
}

func (db *QPDataBase) FindAllBots() ([]QPBot, error) {
	bots := []QPBot{}
	err := db.Connection.Select(&bots, "SELECT * FROM bots")
	return bots, err
}
