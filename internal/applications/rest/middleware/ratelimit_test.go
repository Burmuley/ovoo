package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFailureLimiter(t *testing.T) {
	l := newFailureLimiter(3, time.Minute)
	now := time.Now()

	for i := 0; i < 3; i++ {
		assert.False(t, l.blocked("1.1.1.1", now))
		l.fail("1.1.1.1", now)
	}
	assert.True(t, l.blocked("1.1.1.1", now))
	assert.False(t, l.blocked("2.2.2.2", now), "other clients are not affected")

	assert.False(t, l.blocked("1.1.1.1", now.Add(time.Minute+time.Second)), "window expired")
}

func TestFailureLimiter_ExpiredEntriesCollected(t *testing.T) {
	l := newFailureLimiter(3, time.Minute)
	now := time.Now()
	l.fail("1.1.1.1", now)
	l.fail("2.2.2.2", now.Add(2*time.Minute))
	assert.Len(t, l.entries, 1)
}

func TestClientKey(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:4321"
	r.Header.Set("X-Forwarded-For", "6.6.6.6")
	assert.Equal(t, "10.0.0.1", clientKey(r))
}
