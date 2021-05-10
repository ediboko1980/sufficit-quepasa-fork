package models

var ENV struct{}

func (_ ENV) IsDevelopment() bool {
	environment, _ := getenvStr("APP_ENV")
	if environment == "development" {
		return true
	} else {
		return false
	}
}