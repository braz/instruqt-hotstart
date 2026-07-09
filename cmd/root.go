package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/eoinbrazil/instruqt-hotstart/instruqt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// v is the shared Viper instance for configuration resolution.
var v = viper.New()

// rootCmd is the base command. Subcommands are registered in their own files.
var rootCmd = &cobra.Command{
	Use:           "instruqt-hotstart",
	Short:         "Create and inspect Instruqt hot start pools",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.String("api-key", "", "Instruqt API key (or env INSTRUQT_API_KEY)")
	pf.String("team", "", "team slug scoping all operations (or env INSTRUQT_TEAM)")
	pf.String("endpoint", instruqt.DefaultEndpoint, "GraphQL endpoint (or env INSTRUQT_ENDPOINT)")
	pf.String("config", "", "config file (default ./config.yaml if present)")
	pf.Bool("json", false, "output JSON instead of a table")

	cobra.OnInitialize(initConfig)
}

// initConfig wires the shared Viper instance from the root command's flags.
func initConfig() {
	configureViper(v, rootCmd.PersistentFlags())
}

// configureViper binds env (INSTRUQT_*) and an optional config file to vp, with
// flags bound so precedence is flag > env > config file > default. A missing
// default config file is not an error.
func configureViper(vp *viper.Viper, flags *pflag.FlagSet) {
	vp.SetEnvPrefix("INSTRUQT")
	vp.AutomaticEnv()
	_ = vp.BindPFlags(flags)

	if cfg := vp.GetString("config"); cfg != "" {
		vp.SetConfigFile(cfg)
	} else {
		vp.SetConfigName("config")
		vp.SetConfigType("yaml")
		vp.AddConfigPath(".")
	}
	_ = vp.ReadInConfig()
}

// Execute runs the root command and exits non-zero on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// signalContext returns a context cancelled on SIGINT with a default timeout.
func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancelSig := signal.NotifyContext(context.Background(), os.Interrupt)
	ctx, cancelTimeout := context.WithTimeout(ctx, 60*time.Second)
	return ctx, func() { cancelTimeout(); cancelSig() }
}

// newClient constructs a client from resolved config, erroring early if the
// API key is missing so we never make a doomed network call.
func newClient() (*instruqt.Client, error) {
	apiKey := v.GetString("api-key")
	if apiKey == "" {
		return nil, fmt.Errorf("no API key: set --api-key or INSTRUQT_API_KEY")
	}
	endpoint := v.GetString("endpoint")
	if endpoint == "" {
		endpoint = instruqt.DefaultEndpoint
	}
	return instruqt.New(apiKey, instruqt.WithEndpoint(endpoint)), nil
}

// teamSlug returns the resolved team slug, erroring if unset.
func teamSlug() (string, error) {
	team := v.GetString("team")
	if team == "" {
		return "", fmt.Errorf("no team: set --team or INSTRUQT_TEAM")
	}
	return team, nil
}
