package simclock

import (
	"sync"
	"time"
)

type SimClock struct {
	mu          sync.RWMutex
	currentHour int64
	startedAt   time.Time
	speed       int
}

func New(initialHour int64, startedAt time.Time, speed int) *SimClock {
	if speed < 1 {
		speed = 1
	}
	return &SimClock{
		currentHour: initialHour,
		startedAt:   startedAt,
		speed:       speed,
	}
}

// Tick advances the clock by 1 simulation hour and returns the new hour.
func (c *SimClock) Tick() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentHour++
	return c.currentHour
}

// Now returns the current simulation hour.
func (c *SimClock) Now() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentHour
}

// Speed returns the current speed multiplier.
func (c *SimClock) Speed() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.speed
}

// SetSpeed updates the speed multiplier.
func (c *SimClock) SetSpeed(s int) {
	if s < 1 {
		s = 1
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.speed = s
}

// ToDisplayTime converts a simulation hour to a displayable time.Time.
func (c *SimClock) ToDisplayTime(simHour int64) time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.startedAt.Add(time.Duration(simHour) * time.Hour)
}

// TickInterval returns the real-time duration between ticks based on speed.
// At 1x speed: 1 tick per second (for practical demo purposes).
// The speed multiplier divides this interval further.
func (c *SimClock) TickInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Second / time.Duration(c.speed)
}
