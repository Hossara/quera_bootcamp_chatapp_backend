package server

import "github.com/spf13/cobra"

func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "HTTP server commands",
	}

	cmd.AddCommand(NewStartServerCommand())

	return cmd
}
