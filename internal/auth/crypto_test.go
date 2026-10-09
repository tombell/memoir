package auth

import (
	"testing"
	"time"
)

func TestPasswordHash(t *testing.T) {
	password := "a sufficiently long password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	other, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	if hash == other {
		t.Fatal("password salt reused")
	}
	if !VerifyPassword(hash, password) || VerifyPassword(hash, "the wrong password") {
		t.Fatal("password verification failed")
	}
	for _, malformed := range []string{"!", "", "$argon2id$v=19$m=999999999,t=3,p=1$bad$bad", "$argon2id$v=19$m=65536,t=0,p=0$bad$bad"} {
		if VerifyPassword(malformed, password) {
			t.Fatal("malformed hash accepted")
		}
	}
	if _, err := HashPassword("short"); err == nil {
		t.Fatal("short password accepted")
	}
}

func TestTokens(t *testing.T) {
	a, err := RandomToken()
	if err != nil {
		t.Fatal(err)
	}
	b, err := RandomToken()
	if err != nil {
		t.Fatal(err)
	}
	if a == b || !ValidToken(a) || ValidToken(a+"=") || ValidToken("short") {
		t.Fatal("invalid random token behavior")
	}
}

func TestLimiterExpiryAndBound(t *testing.T) {
	l := NewLimiter(2, time.Hour)
	now := time.Unix(0, 0)
	l.now = func() time.Time { return now }
	if !l.Allow("a") || !l.Allow("a") || l.Allow("a") {
		t.Fatal("attempt limit not enforced")
	}
	now = now.Add(time.Hour)
	if !l.Allow("a") {
		t.Fatal("expired limit did not clear")
	}
}
