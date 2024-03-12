package utils

import "regexp"

const pattern = `^(([a-zA-Z0-9.-]+:\d{2,5}\/)?[a-z0-9]+(?:[._-][a-z0-9]+)*\/)?[a-z0-9]+(?:[._-][a-z0-9]+)*(?::[a-zA-Z0-9._-]+)?$`

func IsValidTag(input string) bool {
	matched, _ := regexp.MatchString(pattern, input)
	return matched
}
