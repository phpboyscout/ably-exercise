package client

import (
	"github.com/rs/zerolog"
	"os"
)

// StateTTL represents the number of seconds state should be retained for.
const (
	// MaxSeed upper limit for seeding service
	MaxSeed = 0xff

	// MaxQty upper limit for the number of values to be returned
	MaxQty = 0xffff

	// TransportReconnectTimeout in event grpc transport is lost how long, in seconds, to keep trying to reconnect for before giving up
	TransportReconnectTimeout = 300
)

// logger represents a configured instance of zerolog.
var logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
