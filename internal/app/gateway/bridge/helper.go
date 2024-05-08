package bridge

import (
	"bytes"
	"errors"
	"io"
	"regexp"

	"github.com/labstack/echo/v4"
)

// BindBody binds the request body to the given request object.
func BindBody(c echo.Context, req interface{}) error {
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	// Reset the request body for subsequent reads.
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.Bind(req); err != nil {
		return err
	}
	// Reset the body to its original state to be read again by the next handler.
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return nil
}

// ExtractTriggerIdFromHostname extracts the trigger ID from the hostname.
func ExtractTriggerIdFromHostname(hostname string) (string, error) {
	// Compile the regular expression with a capturing group for the 32-char string.
	re, err := regexp.Compile(`^([a-zA-Z0-9]{32})\.func(\.[a-zA-Z0-9-]+)*\.[a-zA-Z0-9-]+\.[a-zA-Z]+$`)
	if err != nil {
		// Handle regex compilation error
		return "", err
	}

	// Use FindStringSubmatch to find matches and capturing groups.
	matches := re.FindStringSubmatch(hostname)
	if matches == nil {
		// No matches found, return an error
		return "", errors.New("invalid hostname format")
	}

	// matches[0] is the entire match, matches[1] is the first capturing group.
	return matches[1], nil
}
