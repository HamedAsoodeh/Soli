package cmd

import (
	"fmt"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate-arabica",
		Aliases: []string{""},
		Short:   "state migration for arabica",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			genPath := fmt.Sprintf("%s/%s", clientCtx.HomeDir, "config/genesis.json")
			app.MigrateGenesisStatev070(genPath)
			return nil
		},
	}
	return cmd
}
