package cmd

import (
	"fmt"
	"strings"

	"github.com/jmsperu/sftpgo/internal/config"
	"github.com/spf13/cobra"
)

var (
	addPort    int
	addKeyFile string
)

var addCmd = &cobra.Command{
	Use:   "add name user@host",
	Short: "Save a connection",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		target := args[1]

		if !strings.Contains(target, "@") {
			return fmt.Errorf("target must be user@host")
		}

		parts := strings.SplitN(target, "@", 2)
		port := 22
		if addPort != 0 {
			port = addPort
		}

		host := parts[1]
		if idx := strings.LastIndex(host, ":"); idx != -1 {
			fmt.Sscanf(host[idx+1:], "%d", &port)
			host = host[:idx]
		}

		conn := config.Connection{
			Host:    host,
			Port:    port,
			User:    parts[0],
			KeyFile: addKeyFile,
		}

		cfg, _ := config.Load(cfgFile)
		if cfg.Connections == nil {
			cfg.Connections = make(map[string]config.Connection)
		}
		cfg.Connections[name] = conn

		if err := config.Save(cfgFile, cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Printf("Saved connection %q (%s@%s:%d)\n", name, conn.User, conn.Host, conn.Port)
		return nil
	},
}

func init() {
	addCmd.Flags().IntVarP(&addPort, "port", "p", 22, "SSH port")
	addCmd.Flags().StringVarP(&addKeyFile, "key", "i", "", "path to private key")
	rootCmd.AddCommand(addCmd)
}
