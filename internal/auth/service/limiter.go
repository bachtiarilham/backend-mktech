package service

import (
	"sync"
	"time"

	"ptmuda.tech/backend/internal/apperror"
)

type attempt struct {
	count     int
	resetTime time.Time
}

type Limiter struct {
	mu       sync.Mutex
	attempts map[string]attempt
}

func NewLimiter() *Limiter {
	l := &Limiter{attempts: make(map[string]attempt)}
	go l.cleanupLoop()
	return l
}

func (l *Limiter) Limit(key string, maxAttempts int, window time.Duration) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	a, ok := l.attempts[key]
	if !ok || now.After(a.resetTime) {
		l.attempts[key] = attempt{count: 1, resetTime: now.Add(window)}
		return nil
	}

	if a.count >= maxAttempts {
		return apperror.ErrRateLimited
	}

	a.count++
	l.attempts[key] = a
	return nil
}

func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for k, a := range l.attempts {
			if now.After(a.resetTime) {
				delete(l.attempts, k)
			}
		}
		l.mu.Unlock()
	}
}
