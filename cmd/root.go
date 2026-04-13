package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "sftpgo",
	Short: "Modern SFTP/SCP client",
	Long: `sftpgo - A modern SFTP/SCP client with saved connections,
interactive shell, progress bars, and directory sync.

Clean alternative to WinSCP/FileZilla as a single binary.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.sftpgo.yml)")
}
