package spotifyauth

import "errors"

var (
	ErrNotPlaying       = errors.New("no track currently playing")
	ErrUnauthorized     = errors.New("unauthorized - invalid or expired token")
	ErrRateLimited      = errors.New("rate limit exceeded")
	ErrInvalidResponse  = errors.New("invalid API response")
)