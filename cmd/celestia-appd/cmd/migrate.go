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

			genPath, err := cmd.Flags().GetString("path")
			if err != nil {
				return err
			}

			newGenPath := fmt.Sprintf("%s/%s", clientCtx.HomeDir, "config/genesis.json")
			return app.MigrateGenesisStatev070(genPath, newGenPath)
		},
	}
	cmd.Flags().String("path", "~/new_genesis.json", "specify the path to the already saved json")
	return cmd
}
