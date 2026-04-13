package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jmsperu/sftpgo/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List saved connections",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No saved connections. Use 'sftpgo add' to save one.")
				return nil
			}
			return err
		}

		if len(cfg.Connections) == 0 {
			fmt.Println("No saved connections. Use 'sftpgo add' to save one.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tHOST\tUSER\tPORT\tKEY")
		for name, c := range cfg.Connections {
			key := "-"
			if c.KeyFile != "" {
				key = c.KeyFile
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n", name, c.Host, c.User, c.Port, key)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
