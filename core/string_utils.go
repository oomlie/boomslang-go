package core

import "strings"

type indentlevel struct {
	tabs   int
	spaces int
}

func TrimIndent(s string) (string, indentlevel) {
	count := indentlevel{}
	for {
		if strings.HasPrefix(s, "\t") {
			s = strings.TrimPrefix(s, "\t")
			count.tabs += 1
		} else if strings.HasPrefix(s, " ") {
			s = strings.TrimPrefix(s, " ")
			count.spaces += 1
		} else {
			break
		}
	}
	return strings.TrimSpace(s), count
}
func FirstRune(s string) rune {
	for _, r := range s {
		return r
	}
	return 0
}
