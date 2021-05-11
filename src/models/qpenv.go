package models

type Environment struct{}

var ENV Environment

func (_ *Environment) IsDevelopment() bool {
	environment, _ := getenvStr("APP_ENV")
	if environment == "development" {
		return true
	} else {
		return false
	}
}
