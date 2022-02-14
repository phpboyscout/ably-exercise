package grpc

import (
	"context"
	v1 "exercise/pkg/ably/v1"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"math/big"
	"time"
)

type Stream interface {
	Recv() (*v1.Response, error)
	grpc.ClientStream
}

type ClientInterface interface {
	Context() context.Context
	Connection() *grpc.ClientConn
	Client() v1.ServiceClient
	Retry() chan big.Int
	Done() chan bool
	Cancel()
	Reconnect() bool
}

// Client struct to encapsulate interactions with the grpc services
type Client struct {
	ctx        context.Context
	cancFunc   context.CancelFunc
	connection *grpc.ClientConn
	client     v1.ServiceClient
	retry      chan big.Int
	done       chan bool
}

// buildClientConnection creates a grpc connection based on teh supplied flags.
// includes some simple retry logic in the event of connection loss
// todo: implement TLS
func buildClientConnection(flags *pflag.FlagSet) *grpc.ClientConn {
	server, err := flags.GetString("dsn")
	if err != nil {
		logger.Fatal().Err(err)
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithKeepaliveParams(kacp))

	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponentialWithJitter(100*time.Millisecond, 0.10)),
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted, codes.Canceled, codes.Unavailable, codes.Unknown),
	}

	opts = append(opts, grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)))

	conn, err := grpc.Dial(server, opts...)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	return conn
}

// buildClientContext builds a configured context and cancel func based on the supplied flags .
func buildClientContext(flags *pflag.FlagSet) (context.Context, context.CancelFunc) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	if stateless, err := flags.GetBool("stateless"); stateless && err == nil {
		return ctx, cancelFunc
	}

	var clientID = uuid.New().String()
	if cid, err := flags.GetString("client-id"); err == nil && cid != "" {
		clientID = cid
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "client-id", clientID)
	logger.Debug().Msgf("Client ID: %s", clientID)

	return ctx, cancelFunc
}

// Reconnect attempts to reconnect transport in the event of failure,
// returns true if reconnection successful false if timeout reached.
// In the event of a successful reconnection otherwise false.
// todo: needs expansion to cover more scenarios
func (c *Client) Reconnect() bool {
	ctx, cancel := context.WithTimeout(c.Context(), TransportReconnectTimeout*time.Second)
	defer cancel()

	conn := c.Connection()

	state := conn.GetState()
	waiting := true

	// assuming we have lost transport/connection we will get an idle state, so attempt to reconnect
	if state != connectivity.Ready && state != connectivity.Connecting {
		conn.Connect()
	}

	for state != connectivity.Ready && waiting {
		waiting = conn.WaitForStateChange(ctx, state)
		state = conn.GetState()
		if state != connectivity.Ready && state != connectivity.Connecting {
			conn.Connect()
		}
	}

	if !waiting {
		return true
	}

	return false
}

// Context returns the current context
func (c *Client) Context() context.Context {
	return c.ctx
}

// Connection returns the currently configured connection
func (c *Client) Connection() *grpc.ClientConn {
	return c.connection
}

// Client returns the current client
func (c *Client) Client() v1.ServiceClient {
	return c.client
}

// Retry returns the channel for tracking of retries
func (c *Client) Retry() chan big.Int {
	return c.retry
}

// Done returns the channel for tracking completion of a request
func (c *Client) Done() chan bool {
	return c.done
}

// Cancel triggers the cancel func for the current context
func (c *Client) Cancel() {
	c.cancFunc()
}

// NewClient creates a new configured grpc based on supplied flags.
func NewClient(flags *pflag.FlagSet) *Client {
	ctx, cancFunc := buildClientContext(flags)
	conn := buildClientConnection(flags)
	return &Client{
		ctx:        ctx,
		cancFunc:   cancFunc,
		connection: conn,
		client:     v1.NewServiceClient(conn),
		retry:      make(chan big.Int),
		done:       make(chan bool),
	}
}
