package config

import "os"

// Get returns value for key or fallback when unset/empty.
func Get(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
