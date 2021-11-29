package main

import (
	"reflect"
	"sort"
	"strings"
)

func eq(a, b string) bool {
	// log.Printf("%v eq %v", a, b)
	return a == b
}

func Equals(a, b []string) bool {
	if len(a)-1 != len(b) {
		return false
	}
	if len(a) == 0 && len(b) == 0 {
		return true
	}

	n := []string{}
	for _, v := range a {
		if strings.Contains(v, "PATH=") {
			continue
		}
		n = append(n, v)
	}
	sort.Strings(n)
	sort.Strings(b)

	// found := false
	// for k, v := range n {
	// 	if !eq(v, b[k]) {
	// 		found = true
	// 	}
	// }

	return reflect.DeepEqual(n, b)
}
