package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	// failedAuthLimit is the number of failed Basic authentications allowed per client
	// address within failedAuthWindow.
	failedAuthLimit  = 10
	failedAuthWindow = time.Minute
)

// failureLimiter counts failed authentication attempts per client address in a fixed window.
// Per-account lockout can not slow down guessing of logins that do not exist, so this
// complements it. It is in-memory and per process.
type failureLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string]*failureEntry
	nextGC  time.Time
}

type failureEntry struct {
	count   int
	resetAt time.Time
}

func newFailureLimiter(limit int, window time.Duration) *failureLimiter {
	return &failureLimiter{limit: limit, window: window, entries: map[string]*failureEntry{}}
}

// clientKey returns the host part of the peer address. Forwarding headers are deliberately
// not trusted since any client can forge them.
func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

// blocked reports whether the client has used up its failed attempts in the current window.
func (l *failureLimiter) blocked(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[key]
	if !ok || !now.Before(e.resetAt) {
		return false
	}

	return e.count >= l.limit
}

// fail records a failed attempt for the client.
func (l *failureLimiter) fail(key string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// drop expired entries from time to time so the map can not grow without bound
	if now.After(l.nextGC) {
		for k, e := range l.entries {
			if !now.Before(e.resetAt) {
				delete(l.entries, k)
			}
		}
		l.nextGC = now.Add(l.window)
	}

	e, ok := l.entries[key]
	if !ok || !now.Before(e.resetAt) {
		e = &failureEntry{resetAt: now.Add(l.window)}
		l.entries[key] = e
	}
	e.count++
}

var basicAuthLimiter = newFailureLimiter(failedAuthLimit, failedAuthWindow)
