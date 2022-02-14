package main

import (
	"crypto/rand"
	"exercise/internal/client"
	"math/big"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	DefaultQty  *big.Int
	DefaultSeed = big.NewInt(0)

	// logger represents a configured instance of zerolog.
	logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "Start the ably distributed exercise grpc",
	Long: `Start the ably distributed exercise grpc

An implementation as defined at https://gist.github.com/mattheworiordan/3f2f45ce1f6689c249c4195f38f1b6b7
`,
	Run: func(cmd *cobra.Command, _ []string) {
		cmd.Help()
	},
}

// rootCmd represents the base command when called without any subcommands.
var doublerCmd = &cobra.Command{
	Use:     "doubler",
	Example: "client doubler -d localhost:9090 -n 10 -a 1",
	Short:   "Run the stateless client (doubler)",
	Long: `Run the stateless client (doubler) to generate a sequence of values where they double in value

An implementation as defined at https://gist.github.com/mattheworiordan/3f2f45ce1f6689c249c4195f38f1b6b7
`,
	Run: func(cmd *cobra.Command, _ []string) {
		client := client.NewClient(cmd.Flags())
		if err := client.Start("doubler"); err != nil {
			logger.Error().Err(err).Msg("An unhandled error occurred")
		}
	},
}

// rootCmd represents the base command when called without any subcommands.
var randomCmd = &cobra.Command{
	Use:     "random",
	Example: "client doubler -d localhost:9090 -n 10",
	Short:   "Run the stateful client (random)",
	Long: `Run the stateful client (doubler) to generate a sequence of random values less than 0xfff

An implementation as defined at https://gist.github.com/mattheworiordan/3f2f45ce1f6689c249c4195f38f1b6b7
`,
	Run: func(cmd *cobra.Command, _ []string) {
		client := client.NewClient(cmd.Flags())
		if err := client.Start("random"); err != nil {
			logger.Error().Err(err).Msg("An unhandled error occurred")
		}
	},
}

func main() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	DefaultQty, _ = rand.Int(rand.Reader, big.NewInt(client.MaxQty))
	DefaultSeed, _ = rand.Int(rand.Reader, big.NewInt(client.MaxSeed))
	rootCmd.AddCommand(doublerCmd)
	rootCmd.AddCommand(randomCmd)
	rootCmd.PersistentFlags().StringP("dsn", "d", "localhost:9090", "the server and port that the grpc should connect to")
	rootCmd.PersistentFlags().Int64P("qty", "n", DefaultQty.Int64(), "override the RNG for how many values should be returned")
	doublerCmd.Flags().Int64P("seed", "a", DefaultSeed.Int64(), "anything other than zero overrides the RNG for the seed value")
	randomCmd.Flags().BoolP("stateless", "s", false, "run the grpc as stateless")
	randomCmd.Flags().Int64P("last", "l", 0, "the last value seen by the client")
	randomCmd.Flags().StringP("client-id", "c", "", "manually ser the client-id to use")
}
