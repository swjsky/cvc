package utils

import (
	"strings"
)

//GlobAsterisk - * pattern match
func GlobAsterisk(pattern, s string, CaseMind bool) (con bool) {
	if pattern == "" {
		return false
	}
	if !CaseMind {
		pattern = strings.ToLower(pattern)
		s = strings.ToLower(s)
	}
	con = true
	tmp := s
	arr := strings.Split(pattern, "*")

	for _, substr := range arr {
		p := strings.Index(tmp, substr)
		if p == -1 {
			con = false
			return
		}
		tmp = tmp[p+len(substr):]
	}

	if !strings.HasPrefix(pattern, "*") && !strings.HasPrefix(s, arr[0]) {
		con = false
		return
	}

	if !strings.HasSuffix(pattern, "*") && !strings.HasSuffix(s, arr[len(arr)-1]) {
		con = false
		return
	}
	return
}
