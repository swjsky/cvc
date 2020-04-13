package utils

import "unicode"

//Sanitize - remove Symbol/Punct/Control charactors from string
func Sanitize(s string, keepList ...rune) string {
	result := make([]rune, 0, len(s))
	for _, c := range []rune(s) {
		switch {
		case in(c, keepList...):
			goto keep
		case unicode.IsPunct(c):
			continue
		case unicode.IsSymbol(c):
			continue
		case unicode.IsControl(c):
			continue
		}
	keep:
		result = append(result, c)
	}
	return string(result)
}

func in(r rune, rs ...rune) bool {
	for _, c := range rs {
		if c == r {
			return true
		}
	}
	return false
}
