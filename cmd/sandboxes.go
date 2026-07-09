package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "sandboxes",
		Short: "List sandboxes for a team (use their IDs with create --sandbox-ids)",
		RunE:  runSandboxes,
	}
	rootCmd.AddCommand(c)
}

func runSandboxes(cmd *cobra.Command, _ []string) error {
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

	sandboxes, err := client.Sandboxes(ctx, team)
	if err != nil {
		return err
	}
	asJSON, _ := cmd.Flags().GetBool("json")
	return renderSandboxes(cmd.OutOrStdout(), sandboxes, asJSON)
}
