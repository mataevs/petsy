package utils

import (
//"regexp"
)

const RegexEmail = `^[a-z0-9._\-+]+@[a-z0-9.\-]+\.[a-z]{2,4}$`

func IsEmailAddress(email string) bool {
	// exp, _ := regexp.Compile(RegexEmail)

	// return exp.MatchString(email)
	return true
}

func IsEmpty(value interface{}) bool {
	switch value.(type) {
	case string:
		return value == ""
	case int:
		return value == 0
	default:
		return value == nil
	}
}
