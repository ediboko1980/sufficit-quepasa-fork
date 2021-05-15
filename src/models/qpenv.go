package models

import (
	"errors"
	"os"
	"strconv"
)

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

func (_ *Environment) DEBUGRequests() bool {

	if ENV.IsDevelopment() {
		environment, err := GetEnvBool("DEBUGREQUESTS", true)
		if err == nil {
			return environment
		}
	}

	return false
}

func (_ *Environment) DEBUGJsonMessages() bool {

	if ENV.IsDevelopment() {
		environment, err := GetEnvBool("DEBUGJSONMESSAGES", true)
		if err == nil {
			return environment
		}
	}

	return false
}

var ErrEnvVarEmpty = errors.New("getenv: environment variable empty")

func GetEnvBool(key string, value bool) (bool, error) {
	result := value
	s, err := getenvStr(key)
	if err == nil {
		trying, err := strconv.ParseBool(s)
		if err == nil {
			result = trying
		}
	}
	return result, err
}

func getenvStr(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return v, ErrEnvVarEmpty
	}
	return v, nil
}
