package clock

import "time"

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

type FakeClock struct {
	Current time.Time
}

func (f FakeClock) Now() time.Time {
	return f.Current
}
