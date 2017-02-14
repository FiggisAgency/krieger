package utils

import (
	"fmt"
	"regexp"
)

//MatchRegex puts all matches into a map
//If the pattern is nil, or the number of subexpressions is zero
//the function will return nil
func MatchRegex(pattern *regexp.Regexp, src string) map[string]string {
	if pattern == nil {
		return nil
	}
	i := pattern.NumSubexp()
	if i <= 0 {
		return nil
	}

	m := make(map[string]string, i)

	for _, s := range pattern.SubexpNames() {
		if s == "" {
			continue
		}
		m[s] = pattern.ReplaceAllString(src, fmt.Sprintf("${%s}", s))
	}
	return m
}
