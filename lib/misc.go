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

func TypeSymbolForThing(thing interface{}) (ts string) {
	ts = "?"
	switch typed := thing.(type) {
	case *State:
		ts = "S"
		if typed == nil {
			ts += "?"
		}
	case *Group:
		ts = "G"
		if typed == nil {
			ts += "?"
		}
	case *NFA:
		ts = "N"
		if typed == nil {
			ts += "?"
		}
	}
	return
}

func (un *numberedItems) onlyOnce(thing interface{}) bool {
	if un.in(thing) {
		return false
	}
	un.get(thing)
	return true
}

func (un *numberedItems) get(thing interface{}) string {
	ts := TypeSymbolForThing(thing)
	if ts[len(ts)-1] == '?' {
		return ts
	}
	if un.id[ts] == nil {
		un.id[ts] = make(map[interface{}]int)
	}
	if un.id[ts][thing] == 0 {
		un.next[ts] += 1
		un.id[ts][thing] = un.next[ts]
	}
	return fmt.Sprintf("%s%d", ts, un.id[ts][thing])
}

func (un *numberedItems) in(thing interface{}) bool {
	ts := TypeSymbolForThing(thing)
	if thing == nil {
		return false
	}
	if un.id[ts] == nil {
		return false
	}
	if un.id[ts][thing] == 0 {
		return false
	}
	return true
}

func makeNumberedItems() *numberedItems {
	return &numberedItems{id: make(map[string]map[interface{}]int),
		next: make(map[string]int)}
}

var GLOBAL_UN *numberedItems = makeNumberedItems()

func GetTagList[T any](things []T) string {
	var ret []string
	for _, thing := range things {
		ret = append(ret, GetTag(thing))
	}
	return strings.Join(ret, ", ")
}

func GetTag(thing interface{}) string {
	return GLOBAL_UN.get(thing)
}

func TagDefined(thing interface{}) bool {
	return GLOBAL_UN.in(thing)
}

func GetFTagList[T any](things []T) string {
	var ret []string
	for _, thing := range things {
		ret = append(ret, GetFTag(thing))
	}
	return strings.Join(ret, ", ")
}

func GetFTag(item interface{}) (t string) {
	t = GetTag(item)
	switch t {
	case "N?":
		t = "F"
	case "S?":
		t = "ε"
	}
	return
}
