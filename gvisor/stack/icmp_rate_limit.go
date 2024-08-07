package stack

import "golang.org/x/time/rate"

// ICMPRateLimiter is a global rate limiter that controls the generation of
// ICMP messages generated by the stack.
type ICMPRateLimiter struct {
	*rate.Limiter
}
