package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "get",
		Short: "Get a hot start pool by ID",
		RunE:  runGet,
	}
	c.Flags().String("id", "", "hot start pool ID (required)")
	_ = c.MarkFlagRequired("id")
	rootCmd.AddCommand(c)
}

func runGet(cmd *cobra.Command, _ []string) error {
	id, _ := cmd.Flags().GetString("id")
	if id == "" {
		return fmt.Errorf("--id is required")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	ctx, cancel := signalContext()
	defer cancel()

	pool, err := client.HotStartPool(ctx, id)
	if err != nil {
		return err
	}
	asJSON, _ := cmd.Flags().GetBool("json")
	return renderPool(cmd.OutOrStdout(), pool, asJSON)
}
