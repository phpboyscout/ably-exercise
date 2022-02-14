package main

import (
	"errors"
	"exercise/internal/service"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	v1 "exercise/pkg/ably/v1"
)

var (
	ErrPortRequired = errors.New("you must provide a port number to start the server")
	ErrPortNumber   = errors.New("the first arg must be a valid port number")
)

// logger represents a configured instance of zerolog.
var logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     "server",
	Example: "server 9090",
	Short:   "Start the ably distributed exercise server",
	Long:    `Start the ably distributed exercise server`,
	Args:    validArgs,
	Run:     runServer,
}

// validArgs ensures that the first positional argument passed to the command is a valid port number.
func validArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return ErrPortRequired
	}

	if _, err := strconv.Atoi(args[0]); err != nil {
		return ErrPortNumber
	}

	return nil
}

// buildServer creates a configured instance of a grpc server.
func buildServer() *grpc.Server {
	var opts []grpc.ServerOption

	// todo: implement TLS
	opts = append(opts, grpc.Creds(insecure.NewCredentials()))
	srv := grpc.NewServer(opts...)

	service := service.NewService()
	go service.MaintainStates() // todo: interim solution - needs better handling

	v1.RegisterServiceServer(srv, service)

	return srv
}

// runServer starts a grpc server based upon the supplied arguments and flags.
func runServer(cmd *cobra.Command, args []string) {
	port, err := strconv.Atoi(args[0])
	if err != nil {
		logger.Fatal().Err(err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Fatal().Err(err)
	}

	srv := buildServer()

	logger.Info().Msgf("Starting server on port %d", port)
	if err := srv.Serve(listener); err != nil {
		logger.Fatal().Err(err)
	}
	logger.Info().Msg("Stopped server")
}

func main() {
	cobra.CheckErr(rootCmd.Execute())
}
