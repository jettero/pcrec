package lib

import (
	"fmt"
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

func uniqueStrings(strings []string) []string {
	uniqueMap := make(map[string]bool)
	var uniqueSlice []string

	for _, str := range strings {
		if !uniqueMap[str] {
			uniqueSlice = append(uniqueSlice, str)
			uniqueMap[str] = true
		}
	}

	return uniqueSlice
}

type numberedItems struct {
	id   map[string]map[interface{}]int
	next map[string]int
}

func (un *numberedItems) get(tag string, thing interface{}) string {
	if thing == nil {
		return fmt.Sprintf("%s%d", tag, 0)
	}
	if un.id[tag] == nil {
		un.id[tag] = make(map[interface{}]int)
	}
	if un.id[tag][thing] == 0 {
		un.next[tag] += 1
		un.id[tag][thing] = un.next[tag]
	}
	return fmt.Sprintf("%s%d", tag, un.id[tag][thing])
}

func (un *numberedItems) in(tag string, thing interface{}) bool {
	if thing == nil {
		return false
	}
	if un.id[tag] == nil {
		return false
	}
	if un.id[tag][thing] == 0 {
		return false
	}
	return true
}

func makeNumberedItems() *numberedItems {
	return &numberedItems{id: make(map[string]map[interface{}]int),
		next: make(map[string]int)}
}
