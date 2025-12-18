package cmd

import (
	serverCmd "github.com/Hossara/quera_bootcamp_chatapp_backend/cmd/server"
	systemCmd "github.com/Hossara/quera_bootcamp_chatapp_backend/cmd/system"
	"github.com/spf13/cobra"
)

var (
	configPath string
)

var rootCmd = &cobra.Command{
	Use:   "chatapp-server",
	Short: "Chat Application Backend Server",
	Long:  `A real-time chat application backend server with WebSocket support`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global config flag, available for all commands.
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", ".", "Path to config file directory")

	// Attach top-level command trees.
	rootCmd.AddCommand(systemCmd.NewSystemCommand())
	rootCmd.AddCommand(serverCmd.NewServerCommand())
}
