package cmd

import (
	"fmt"
	"strings"

	"github.com/jmsperu/sftpgo/internal/config"
	"github.com/jmsperu/sftpgo/internal/transfer"
	"github.com/spf13/cobra"
)

var syncDelete bool

var syncCmd = &cobra.Command{
	Use:   "sync ./local server:/remote",
	Short: "Sync local directory to remote",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		local := args[0]
		remote := args[1]

		idx := strings.Index(remote, ":")
		if idx == -1 {
			return fmt.Errorf("invalid remote path %q — use server:/path format", remote)
		}
		name := remote[:idx]
		remotePath := remote[idx+1:]
		if remotePath == "" {
			remotePath = "."
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		connCfg, ok := cfg.Connections[name]
		if !ok {
			return fmt.Errorf("unknown connection %q", name)
		}

		return transfer.Sync(connCfg, local, remotePath, syncDelete)
	},
}

func init() {
	syncCmd.Flags().BoolVar(&syncDelete, "delete", false, "delete remote files not in local")
	rootCmd.AddCommand(syncCmd)
}
