package client

import (
	"errors"
	"exercise/internal/random"
	"fmt"
	"io"
	"math/big"

	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/spf13/pflag"

	"exercise/internal/doubler"
	"exercise/internal/grpc"
	"exercise/internal/state"
)

type Client struct {
	StreamFunc func(state.Stateful) (grpc.Stream, error)
	grpc.ClientInterface
	State state.Stateful
}

func (c *Client) getStream(service string) (grpc.Stream, error) {
	switch service {
	case "random":
		return c.Client().Random(c.Context(), random.GetRequest(c.State), grpcRetry.WithMax(5))
	default:
		return c.Client().Doubler(c.Context(), doubler.GetRequest(c.State), grpcRetry.WithMax(5))
	}
}

// handleStream processes the incoming stream of numbers maintaining an internal client side state
func (c *Client) handleStream(service string, prevSum big.Int) {
	stream, streamErr := c.getStream(service)
	if streamErr != nil {
		logger.Fatal().Err(streamErr).Msg("Unable to open stream")
	}

	checksum := new(big.Int)

	switch service {
	case "doubler":
		logger.Debug().
			Str("tally", c.State.Total().String()).
			Str("value", c.State.Last().String()).
			Send()
	}
	if prevSum.Cmp(big.NewInt(0)) >= 0 {
		checksum.Set(&prevSum)
	}
	for {
		response, err := stream.Recv()
		if err == nil && response.Checksum != nil {
			checksum.SetBytes(response.Checksum)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				c.Done() <- c.State.Total().Cmp(checksum) == 0
			} else {
				logger.Error().Err(err).Send()
				c.Retry() <- *c.State.Total()
			}

			break
		}
		if err == nil && response != nil {
			value := big.Int{}
			value.SetBytes(response.Value)
			c.State.Add(&value)

			l := logger.Debug().Str("tally", c.State.Total().String())
			if response.Checksum != nil {
				l.Str("checksum", checksum.String())
			}
			if response.Value != nil {
				l.Str("value", c.State.Last().String())
			}
			l.Send()
		}
	}
}

// Start begins processing the request and subsequent stream of data
func (c *Client) Start(service string) error {
	defer func() { _ = c.Connection().Close() }()
	go c.handleStream(service, *big.NewInt(0))
	for {
		select {
		case checksum := <-c.Retry():
			logger.Debug().Timestamp().Msg("Stream lost: reconnecting")
			if c.Reconnect() {
				return fmt.Errorf("timed out after %d seconds: %w", TransportReconnectTimeout, errors.New("unable to reconnect"))
			}
			go c.handleStream(service, checksum)
		case success := <-c.Done():
			c.Cancel()
			logger.Info().Msgf("Total: %d (checksum=%t)", c.State.Total(), success)

			return nil
		}
	}
}

// NewClient creates a new configured grpc based on supplied flags.
func NewClient(flags *pflag.FlagSet) *Client {
	var seed = int64(0)

	qty, err := flags.GetInt64("qty")
	if err != nil {
		logger.Fatal().Err(err)
	}

	if flags.Lookup("seed") != nil {
		seed, err = flags.GetInt64("seed")
		if err != nil {
			logger.Fatal().Err(err).Send()
		}
	}

	return &Client{
		ClientInterface: grpc.NewClient(flags),
		State:           state.NewState(qty, []*big.Int{big.NewInt(seed)}),
	}
}
