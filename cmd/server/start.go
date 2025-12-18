package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"github.com/Hossara/quera_bootcamp_chatapp_backend/config"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/repository/ent"
	"github.com/Hossara/quera_bootcamp_chatapp_backend/internal/server"
	"github.com/spf13/cobra"
)

func NewStartServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the chat application backend server",
		Long:  `Start the chat application backend server with WebSocket support`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get config path from flag
			configPath, err := cmd.Root().PersistentFlags().GetString("config")
			if err != nil {
				return fmt.Errorf("failed to read config flag: %w", err)
			}

			// Load configuration
			cfg, err := config.ReadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			log.Printf("Config loaded from: %s", configPath)

			// Initialize database connection
			client, err := ent.Open("postgres", cfg.PostgresDB.DSN())
			if err != nil {
				return fmt.Errorf("failed opening connection to postgres: %w", err)
			}
			defer func() {
				if err := client.Close(); err != nil {
					log.Printf("Error closing database connection: %v", err)
				}
			}()

			// Run database migrations
			log.Println("Running database migrations...")
			if err := client.Schema.Create(context.Background()); err != nil {
				return fmt.Errorf("failed creating schema resources: %w", err)
			}
			log.Println("Database migrations completed successfully")

			// Create and configure server
			srv, err := server.New(cfg, client)
			if err != nil {
				return fmt.Errorf("failed to create server: %w", err)
			}

			// Start server in a goroutine
			serverErrors := make(chan error, 1)
			go func() {
				serverErrors <- srv.Start()
			}()

			// Wait for interrupt signal or server error
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

			select {
			case err := <-serverErrors:
				return fmt.Errorf("server error: %w", err)
			case sig := <-quit:
				log.Printf("Received signal: %v", sig)
			}

			// Graceful shutdown
			if err := srv.Shutdown(); err != nil {
				return fmt.Errorf("server shutdown error: %w", err)
			}

			return nil
		},
	}
}
