package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "configs",
		Short: "List sandbox configs for a team (use their IDs with create --configs)",
		RunE:  runConfigs,
	}
	rootCmd.AddCommand(c)
}

func runConfigs(cmd *cobra.Command, _ []string) error {
	team, err := teamSlug()
	if err != nil {
		return err
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	ctx, cancel := signalContext()
	defer cancel()

	configs, err := client.SandboxConfigs(ctx, team)
	if err != nil {
		return err
	}
	asJSON, _ := cmd.Flags().GetBool("json")
	return renderConfigs(cmd.OutOrStdout(), configs, asJSON)
}
