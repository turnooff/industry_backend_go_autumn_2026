package main

import (
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	newLimiter := &Limiter{}
	newLimiter.clock = clock
	newLimiter.rate = ratePerSec
	newLimiter.burst = burst
	newLimiter.tokens = float64(burst)

	if clock != nil {
		newLimiter.last = clock.Now()
	}

	return newLimiter
}
func (l *Limiter) AllowN(n int) bool {

	l.mu.Lock()
	defer l.mu.Unlock()

	if n <= 0 || n > l.burst || l.burst <= 0 || l.clock == nil {
		return false
	}

	now := l.clock.Now()

	diff := now.Sub(l.last)
	if diff > 0 && l.rate > 0 {
		l.tokens += diff.Seconds() * l.rate
		l.last = now
	}

	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}

	if float64(n) > l.tokens {
		return false
	}

	l.tokens -= float64(n)
	return true
}
