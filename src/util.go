package main

import (
	"os"
	"regexp"
	"strings"
)

func FailError(err error) {
	if err != nil {
		ColorPrint(ERROR, err.Error())
	}
}

func getValueOf(key, fallback string) string {
	value, err := os.ReadFile("/etc/secrets/" + key)
	if err != nil {
		return fallback
	}
	data := string(value)
	data = strings.Trim(data, "\"")
	return data
}

var rxUserRequiresToken = regexp.MustCompile("[a-z0-9]+@[a-z0-9]+![a-z0-9]+")

func userRequiresAPIToken(userID string) bool {
	return rxUserRequiresToken.MatchString(userID)
}
