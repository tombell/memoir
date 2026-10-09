package auth

import (
	"sync"
	"time"
)

type attempt struct {
	count   int
	expires time.Time
}

// Limiter bounds both the number of attempts and the number of stored keys.
// Each process has its own limits. Proxy headers are not trusted implicitly.
type Limiter struct {
	mu      sync.Mutex
	entries map[string]attempt
	limit   int
	window  time.Duration
	now     func() time.Time
}

func NewLimiter(limit int, window time.Duration) *Limiter {
	return &Limiter{entries: make(map[string]attempt), limit: limit, window: window, now: time.Now}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	entry, found := l.entries[key]
	if !found || !entry.expires.After(now) {
		if len(l.entries) >= 4096 {
			for key, entry := range l.entries {
				if !entry.expires.After(now) {
					delete(l.entries, key)
				}
			}
			if len(l.entries) >= 4096 {
				return false
			}
		}
		entry = attempt{expires: now.Add(l.window)}
	}
	if entry.count >= l.limit {
		return false
	}
	entry.count++
	l.entries[key] = entry
	return true
}
