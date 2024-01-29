package utils

import "os"

// SetEnvVariables sets variables to the os environment.
func SetEnvVariables(variables map[string]string) error {
	for key, value := range variables {
		err := os.Setenv(key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetEnvOrElse gets the value from the os environment or use the else value if variable is not present.
func GetEnvOrElse(name, orElse string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return orElse
}
