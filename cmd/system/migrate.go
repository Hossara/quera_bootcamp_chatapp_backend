package system

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Hossara/quera_bootcamp_chatapp_backend/config"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/pkg/constants"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/pkg/database"
	"github.com/spf13/cobra"
)

func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, err := cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				return fmt.Errorf("failed to get config flag: %w", err)
			}

			cfg := config.MustReadConfig(filepath.Dir(cfgPath))

			client, err := database.NewEntClient(cfg.PostgresDB)
			if err != nil {
				return fmt.Errorf("failed to create ent client: %w", err)
			}
			defer client.Close()

			timeout := time.Duration(constants.MigrationTimeoutSeconds) * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if err := database.MigrateEnt(ctx, client); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}

			fmt.Println("Migrations executed successfully.")
			return nil
		},
	}

	return cmd
}
