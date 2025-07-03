package cmd

import (
	"fmt"

	log "github.com/SKevo18/go-simplelog"
	"github.com/spf13/cobra"
)

var configFilePath string

func init() {
	updateCmd.Flags().StringVarP(&configFilePath, "config", "c", "updater.ini", "Path to dependencies file")
	rootCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update [download_into]",
	Short: "Updates server jar and plugins defined in config file. If no plugins or server jar are present, downloads them.",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		downloadInto := "."
		if len(args) > 0 {
			downloadInto = args[0]
		}

		log.Debug(fmt.Sprintf("Downloading into: %s", downloadInto))
		log.Task("Updating server")
		return nil
	},
}
