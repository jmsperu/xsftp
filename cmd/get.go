package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jmsperu/sftpgo/internal/config"
	"github.com/jmsperu/sftpgo/internal/transfer"
	"github.com/spf13/cobra"
)

var getResume bool

var getCmd = &cobra.Command{
	Use:   "get server:/remote/path ./local",
	Short: "Download file or directory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		remote := args[0]
		local := args[1]

		name, remotePath, err := parseRemotePath(remote)
		if err != nil {
			return err
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		connCfg, ok := cfg.Connections[name]
		if !ok {
			return fmt.Errorf("unknown connection %q", name)
		}

		return transfer.Download(connCfg, remotePath, local, getResume)
	},
}

func parseRemotePath(s string) (name, path string, err error) {
	idx := strings.Index(s, ":")
	if idx == -1 {
		return "", "", fmt.Errorf("invalid remote path %q — use server:/path format", s)
	}
	name = s[:idx]
	path = s[idx+1:]
	if path == "" {
		path = "."
	}

	// Verify it's a saved connection
	cfg, loadErr := config.Load("")
	if loadErr != nil && !os.IsNotExist(loadErr) {
		return "", "", loadErr
	}
	if _, ok := cfg.Connections[name]; !ok {
		return "", "", fmt.Errorf("unknown connection %q — use 'sftpgo add' to save one", name)
	}

	return name, path, nil
}

func init() {
	getCmd.Flags().BoolVar(&getResume, "resume", false, "resume interrupted transfer")
	rootCmd.AddCommand(getCmd)
}
