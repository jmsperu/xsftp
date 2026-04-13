package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jmsperu/sftpgo/internal/config"
	"github.com/jmsperu/sftpgo/internal/session"
	"github.com/spf13/cobra"
)

var (
	connectPort    int
	connectKeyFile string
)

var connectCmd = &cobra.Command{
	Use:   "connect [user@host | saved-name]",
	Short: "Open interactive SFTP shell",
	Long:  `Connect to an SFTP server and open an interactive shell with tab completion.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		cfg, err := config.Load(cfgFile)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("loading config: %w", err)
		}

		var connCfg config.Connection

		// Check if it's a saved connection name
		if saved, ok := cfg.Connections[target]; ok {
			connCfg = saved
		} else if strings.Contains(target, "@") {
			parts := strings.SplitN(target, "@", 2)
			host := parts[1]
			port := 22
			if connectPort != 0 {
				port = connectPort
			}
			// Check if host has :port
			if idx := strings.LastIndex(host, ":"); idx != -1 {
				fmt.Sscanf(host[idx+1:], "%d", &port)
				host = host[:idx]
			}
			connCfg = config.Connection{
				Host:    host,
				Port:    port,
				User:    parts[0],
				KeyFile: connectKeyFile,
			}
		} else {
			return fmt.Errorf("unknown saved connection %q — use 'sftpgo add' to save one", target)
		}

		if connectPort != 0 {
			connCfg.Port = connectPort
		}
		if connectKeyFile != "" {
			connCfg.KeyFile = connectKeyFile
		}

		return session.Interactive(connCfg)
	},
}

func init() {
	connectCmd.Flags().IntVarP(&connectPort, "port", "p", 0, "SSH port (default 22)")
	connectCmd.Flags().StringVarP(&connectKeyFile, "key", "i", "", "path to private key")
	rootCmd.AddCommand(connectCmd)
}
