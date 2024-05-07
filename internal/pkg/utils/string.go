package utils

import (
	"math/rand"
	"strings"
	"time"
)

// RandomString generates a random string of a given length.
func RandomString(length int, includeUpper bool, includeNumbers bool) string {
	// Seed the random number generator.
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Define the character set.
	charSet := "abcdefghijklmnopqrstuvwxyz"
	if includeUpper {
		charSet += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}
	if includeNumbers {
		charSet += "0123456789"
	}
	var output strings.Builder
	for i := 0; i < length; i++ {
		randomIndex := random.Intn(len(charSet))
		randomChar := charSet[randomIndex]
		output.WriteByte(randomChar)
	}
	return output.String()
}
