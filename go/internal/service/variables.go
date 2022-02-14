package service

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

// StateTTL represents the number of seconds state should be retained for.
const (
	// StateTTL how many seconds should the service maintain state for
	StateTTL = 30

	// MaxSeed upper limit for seeding service
	MaxSeed = 0xff

	// MaxQty upper limit for the number of values to be returned
	MaxQty = 0xffff
)

// logger represents a configured instance of zerolog.
var logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})

// interval defines the period to wait between generating new values in sequence.
var interval = time.Duration(1) * time.Second
