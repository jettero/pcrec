package lib

import (
	"os"
	"strconv"
	"strings"
)

func TruthyEnv(e string) bool {
	s := strings.ToLower(os.Getenv(e))
	if s == "true" || s == "yes" {
		return true
	}
	if num, err := strconv.Atoi(s); err == nil {
		return num != 0
	}
	return false
}
