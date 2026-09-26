package domain

import "time"

// Login lockout (IDN-02.8): five consecutive failures for the same e-mail
// within fifteen minutes block further attempts for fifteen minutes. The lock is
// always temporary.
const (
	MaxLoginFailures   = 5
	LoginFailureWindow = 15 * time.Minute
	LoginLockDuration  = 15 * time.Minute
)

// Lockout decides, from the failures since the last success (oldest first),
// whether attempts are blocked at the instant now, and until when. Only the last
// MaxLoginFailures failures count: they block when they all fit in the window,
// and the lock lasts LoginLockDuration after the last of them.
func Lockout(failures []time.Time, now time.Time) (until time.Time, blocked bool) {
	if len(failures) < MaxLoginFailures {
		return time.Time{}, false
	}
	last := failures[len(failures)-MaxLoginFailures:]
	if last[len(last)-1].Sub(last[0]) > LoginFailureWindow {
		return time.Time{}, false
	}
	until = last[len(last)-1].Add(LoginLockDuration)
	return until, now.Before(until)
}
