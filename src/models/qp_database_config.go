package models

import (
	"fmt"
	"log"
)

type QPDatabaseConfig struct {
	Driver   string
	Host     string
	DataBase string
	Port     string
	User     string
	Password string
	SSL      string
}

func (config *QPDatabaseConfig) GetConnectionString() (connection string) {
	if config.Driver == "mysql" {
		connection = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			config.User, config.Password, config.Host, config.Port, config.DataBase)
	} else if config.Driver == "postgres" {
		connection = fmt.Sprintf("host=%s dbname=%s port=%s user=%s password=%s sslmode=%s",
			config.Host, config.DataBase, config.Port, config.User, config.Password, config.SSL)
	} else if config.Driver == "sqlite3" {
		connection = "quepasa.db?cache=shared&mode=memory"
	} else { log.Fatal("database driver not supported") }
	return
}
