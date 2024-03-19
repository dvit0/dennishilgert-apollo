package utils

import "regexp"

func IsValidTag(input string) bool {
	pattern := `^(([a-zA-Z0-9.-]+:\d{2,5}\/)?[a-z0-9]+(?:[._-][a-z0-9]+)*\/)?[a-z0-9]+(?:[._-][a-z0-9]+)*(?::[a-zA-Z0-9._-]+)?$`
	matched, _ := regexp.MatchString(pattern, input)
	return matched
}

func IsValidRegistryAddress(input string) bool {
	pattern := `^[\w-]+(?:\.[\w-]+)*(?::\d+)?$`
	matched, _ := regexp.MatchString(pattern, input)
	return matched
}
