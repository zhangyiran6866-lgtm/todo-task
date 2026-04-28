package logger

import (
	"strings"
)

const redacted = "[REDACTED]"

func MaskField(key, value string) string {
	lowerKey := strings.ToLower(key)

	switch {
	case strings.Contains(lowerKey, "password"):
		return MaskPassword(value)
	case strings.Contains(lowerKey, "token"), strings.Contains(lowerKey, "authorization"):
		return MaskToken(value)
	case strings.Contains(lowerKey, "email"):
		return MaskEmail(value)
	default:
		return value
	}
}

func MaskPassword(_ string) string {
	return redacted
}

func MaskToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return token
	}
	if len(token) <= 8 {
		return redacted
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}

func MaskEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return email
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" {
		return redacted
	}

	local := parts[0]
	domain := parts[1]

	if len(local) <= 2 {
		return local[:1] + "*@" + domain
	}
	return local[:2] + strings.Repeat("*", len(local)-2) + "@" + domain
}
