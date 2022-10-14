package cmd

import (
	"github.com/celestiaorg/celestia-app/app"
	"github.com/spf13/cobra"
)

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate-arabica",
		Aliases: []string{""},
		Short:   "state migration for arabica",
		RunE: func(cmd *cobra.Command, args []string) error {
			// clientCtx, err := client.GetClientTxContext(cmd)
			// if err != nil {
			// 	return err
			// }

			genPath, err := cmd.Flags().GetString("path")
			if err != nil {
				return err
			}
			app.MigrateGenesisStatev070(genPath)
			return nil
		},
	}
	cmd.Flags().String("path", "~/new_genesis.json")
	return cmd
}
