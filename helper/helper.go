package helper

import "fmt"

// Helper function to get minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


// Masks sensitive values for logs
func MaskSecret(value string) string {
	if len(value) <= 6 {
		return "******"
	}
	return value[:4] + "****" + value[len(value)-2:]
}


func MaskKey(key string) string {
	if len(key) < 6 {
		return "******"
	}
	return fmt.Sprintf("%s****%s", key[:4], key[len(key)-2:])
}
