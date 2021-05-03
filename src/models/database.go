package models

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

type QPDataBase struct{}

var (
	Sync       sync.Once // Objeto de sinaleiro para garantir uma única chamada em todo o andamento do programa
	Connection *sqlx.DB
)

// GetDB returns a database connection for the given
// database environment variables
func GetDB() *sqlx.DB {
	Sync.Do(func() {
		host := os.Getenv("PGHOST")
		database := os.Getenv("PGDATABASE")
		port := os.Getenv("PGPORT")
		user := os.Getenv("PGUSER")
		password := os.Getenv("PGPASSWORD")
		ssl := os.Getenv("PGSSLMODE")
		connection := fmt.Sprintf("host=%s dbname=%s port=%s user=%s password=%s sslmode=%s",
			host, database, port, user, password, ssl)
		dbconn, err := sqlx.Connect("postgres", connection)

		// Tenta realizar a conexão
		if err != nil {
			log.Fatalln(err)
		}

		dbconn.DB.SetMaxIdleConns(500)
		dbconn.DB.SetMaxOpenConns(1000)
		dbconn.DB.SetConnMaxLifetime(30 * time.Second)

		if err != nil {
			log.Fatalln(err)
		}

		// Definindo uma única conexão para todo o sistema
		Connection = dbconn
	})
	return Connection
}
