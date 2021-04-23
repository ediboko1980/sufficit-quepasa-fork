package models

import (
	"errors"
	"os"
	"strconv"
)

var ErrEnvVarEmpty = errors.New("getenv: environment variable empty")

func isDevelopment() bool {
	environment, _ := getenvStr("APP_ENV")
	if environment == "development" {
		return true
	} else {
		return false
	}
}

func getenvStr(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return v, ErrEnvVarEmpty
	}
	return v, nil
}

func GetEnvBool(key string) (bool, error) {
	s, err := getenvStr(key)
	if err != nil {
		return false, err
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return false, err
	}
	return v, nil
}
