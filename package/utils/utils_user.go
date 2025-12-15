package utils

import "strings"

func GetNameFromEmail(email string) string {
	if idx := strings.Index(email, "@"); idx != -1 {
		return email[:idx]
	}
	return email
}
