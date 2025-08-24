package tftp

import "time"

const (
	defaultTimeout     = 1 * time.Second
	defaultMaxAttempts = 5
)

func newBackoff() *backoff {
	return &backoff{
		timeout:     defaultTimeout,
		maxAttempts: defaultMaxAttempts,
		attempt:     1,
	}
}

type backoff struct {
	timeout     time.Duration
	maxAttempts int
	attempt     int
}

func (b *backoff) backoff() {
	time.Sleep(b.timeout)
	b.attempt++
}

func (b *backoff) reset() {
	b.attempt = 1
}

func (b *backoff) getAttempt() int {
	return b.attempt
}

func (b *backoff) getMaxAttempts() int {
	return b.maxAttempts
}
