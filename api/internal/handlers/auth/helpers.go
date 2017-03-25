package auth

// validCredentials validates that the credentials follow some sane rules.
func validCredentials(username, password string) bool {
	return username != "" && password != ""
}
