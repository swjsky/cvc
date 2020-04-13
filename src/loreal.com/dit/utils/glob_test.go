package utils

import (
	"testing"
)

type Test struct {
	Pattern  string
	S        string
	CaseMind bool
	Match    bool
}

func BenchmarkGlobAsterisk(t *testing.B) {
	tests := []Test{
		// 匹配空字符串
		{
			Pattern:  "abc*",
			S:        "abc",
			CaseMind: true,
			Match:    true,
		},
		// 匹配两个型号
		{
			Pattern:  "abc**",
			S:        "abc",
			CaseMind: true,
			Match:    true,
		},
		// 匹配大小写
		{
			Pattern:  "*def",
			S:        "ABCDEF",
			CaseMind: true,
			Match:    false,
		},
		{
			Pattern:  "*def",
			S:        "af",
			CaseMind: true,
			Match:    false,
		},
		{
			Pattern:  "ab*ef",
			S:        "abcdef",
			CaseMind: true,
			Match:    true,
		},
		{
			Pattern: "ab*ef",
			S:       "af",
			Match:   false,
		},
		{
			Pattern: "https://*gobwas.com",
			S:       "http://safe.gobwas.com",
			Match:   false,
		},
		{
			Pattern: "https://*.google.**",
			S:       "https://account.Google.com",
			Match:   true,
		},
		{
			Pattern:  "https://*.google.**",
			S:        "https://google.com",
			CaseMind: true,
			Match:    false,
		},
		{
			Pattern: "https://google.com",
			S:       "https://GOOGLE.com",
			Match:   true,
		},
		{
			Pattern:  "https://google.com",
			S:        "https://Google.com",
			CaseMind: true,
			Match:    false,
		},
	}

	for _, item := range tests {
		ok := GlobAsterisk(item.Pattern, item.S, item.CaseMind)
		if ok != item.Match {
			t.Fatalf("pattern: \"%s\",\tstring: \"%s\",\tresult: %v,\tshould be: %v\r\n", item.Pattern, item.S, ok, item.Match)
		}
	}
}
