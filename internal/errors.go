package internal

import (
	"regexp"
)

func IsNotFoundErr(err error) bool {
	match, _ := regexp.MatchString("[Nn]ot found", err.Error())
	return match
}
