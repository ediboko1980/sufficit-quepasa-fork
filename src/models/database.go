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
	Config     QPDatabaseConfig
	Connection *sqlx.DB
	Store      IQPStore
	User       IQPUser
	Bot        IQPBot
}

var (
	Sync       sync.Once // Objeto de sinaleiro para garantir uma única chamada em todo o andamento do programa
	Connection *sqlx.DB
)

// GetDB returns a database connection for the given
// database environment variables
func GetDB() *sqlx.DB {
	Sync.Do(func() {
		config := GetDBConfig()
		var connection string
		if config.Driver == "mysql" {
			connection = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
				config.User, config.Password, config.Host, config.Port, config.DataBase)
		} else {
			connection = fmt.Sprintf("host=%s dbname=%s port=%s user=%s password=%s sslmode=%s",
				config.Host, config.DataBase, config.Port, config.User, config.Password, config.SSL)
		}
		dbconn, err := sqlx.Connect(config.Driver, connection)

		// Tenta realizar a conexão
		if err != nil {
			log.Println(err)
		}

		dbconn.DB.SetMaxIdleConns(500)
		dbconn.DB.SetMaxOpenConns(1000)
		dbconn.DB.SetConnMaxLifetime(30 * time.Second)

		if err != nil {
			log.Println(err)
		}

		// Definindo uma única conexão para todo o sistema
		Connection = dbconn
	})
	return Connection
}

func GetDatabase() *QPDataBase {
	db := GetDB()
	config := GetDBConfig()
	var istore IQPStore
	var iuser IQPUser
	var ibot IQPBot

	if config.Driver == "postgres" {
		istore = QPStorePostgres{db}
		iuser = QPUserPostgres{db}
		ibot = QPBotPostgres{db}
	} else if config.Driver == "mysql" {
		istore = QPStoreMysql{db}
		iuser = QPUserMysql{db}
		ibot = QPBotMysql{db}
	} else {
		log.Fatal("database driver not supported")
	}

	return &QPDataBase{*config, db, istore, iuser, ibot}
}

func GetDBConfig() *QPDatabaseConfig {
	config := &QPDatabaseConfig{}
	config.Driver = os.Getenv("DBDRIVER")
	config.Host = os.Getenv("DBHOST")
	config.DataBase = os.Getenv("DBDATABASE")
	config.Port = os.Getenv("DBPORT")
	config.User = os.Getenv("DBUSER")
	config.Password = os.Getenv("DBPASSWORD")
	config.SSL = os.Getenv("DBSSLMODE")
	return config
}
