package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "list",
		Short: "List hot start pools for a team",
		RunE:  runList,
	}
	rootCmd.AddCommand(c)
}

func runList(cmd *cobra.Command, _ []string) error {
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

	pools, err := client.HotStartPools(ctx, team)
	if err != nil {
		return err
	}
	asJSON, _ := cmd.Flags().GetBool("json")
	return renderPools(cmd.OutOrStdout(), pools, asJSON)
}
