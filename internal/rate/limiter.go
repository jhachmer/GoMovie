package rate

import (
	"sync"
	"time"
)

type RateLimiter struct {
	Visitors map[string]*Visitor
	mu       sync.Mutex
	rate     int
	window   time.Duration
}

type Visitor struct {
	lastSeen time.Time
	count    int
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		Visitors: make(map[string]*Visitor),
		rate:     rate,
		window:   window,
	}

	go rl.cleanupVisitors()

	return rl
}

func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.Visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.Visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.Visitors[ip]
	now := time.Now()

	if !exists {
		rl.Visitors[ip] = &Visitor{
			lastSeen: now,
			count:    1,
		}
		return true
	}

	if now.Sub(v.lastSeen) > rl.window {
		v.count = 1
		v.lastSeen = now
		return true
	}

	if v.count >= rl.rate {
		return false
	}

	v.count++
	v.lastSeen = now
	return true
}
