package grpc

import (
	"github.com/rs/zerolog"
	"google.golang.org/grpc/keepalive"
	"os"
	"time"
)

// StateTTL represents the number of seconds state should be retained for.
const (
	// TransportReconnectTimeout in event grpc transport is lost how long, in seconds, to keep trying to reconnect for before giving up
	TransportReconnectTimeout = 300
)

// logger represents a configured instance of zerolog.
var logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})

// kacp configuration for keeping a connection alive
var kacp = keepalive.ClientParameters{
	Time:                1 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             time.Second,     // reconnect 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,            // send pings even without active streams
}
