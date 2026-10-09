package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

const passwordMemory = 64 * 1024

// HashPassword stores the salt, algorithm version, and cost with the hash.
func HashPassword(password string) (string, error) {
	if !utf8.ValidString(password) || utf8.RuneCountInString(password) < 15 || utf8.RuneCountInString(password) > 128 {
		return "", fmt.Errorf("password must contain 15 to 128 characters")
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 3, passwordMemory, 1, 32)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=3,p=1$%s$%s", passwordMemory, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}

func VerifyPassword(encoded, password string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" || len(password) > 512 {
		return false
	}
	var memory, iterations uint32
	var parallelism uint8
	if n, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil || n != 3 || parts[3] != fmt.Sprintf("m=%d,t=%d,p=%d", memory, iterations, parallelism) {
		return false
	}
	if memory < 19456 || memory > 256*1024 || iterations < 1 || iterations > 10 || parallelism < 1 || parallelism > 4 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != 16 {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(want) != 32 {
		return false
	}
	got := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, 32)
	return subtle.ConstantTimeCompare(want, got) == 1
}

func RandomToken() (string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(secret), nil
}

func ValidToken(token string) bool {
	if len(token) != 43 {
		return false
	}
	data, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil && len(data) == 32 && base64.RawURLEncoding.EncodeToString(data) == token
}

func tokenHash(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}
