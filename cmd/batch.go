package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jmsperu/sftpgo/internal/config"
	"github.com/jmsperu/sftpgo/internal/session"
	"github.com/spf13/cobra"
)

var batchCmd = &cobra.Command{
	Use:   "batch server command-file",
	Short: "Run batch commands from file",
	Long: `Execute SFTP commands from a file, one per line.
Supported commands: ls, cd, get, put, mkdir, rm, pwd`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		cmdFile := args[1]

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		connCfg, ok := cfg.Connections[name]
		if !ok {
			return fmt.Errorf("unknown connection %q", name)
		}

		f, err := os.Open(cmdFile)
		if err != nil {
			return fmt.Errorf("opening command file: %w", err)
		}
		defer f.Close()

		var commands []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				commands = append(commands, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading command file: %w", err)
		}

		return session.Batch(connCfg, commands)
	},
}

func init() {
	rootCmd.AddCommand(batchCmd)
}
