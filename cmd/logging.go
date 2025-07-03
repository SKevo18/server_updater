package cmd

import (
	log "github.com/gwillem/go-simplelog"
	"github.com/spf13/cobra"
)

var (
	verbosity int
	silent    bool
)

func init() {
	rootCmd.PersistentFlags().IntVarP(&verbosity, "verbosity", "v", 1, "Verbosity level (0=debug, 1=task, 2=warn, 3=alert, 4=error)")
	rootCmd.PersistentFlags().BoolVarP(&silent, "silent", "s", false, "Silent mode - no log output")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		setLogLevel(verbosity)
		silenceLogging(silent)
		return nil
	}
}

func setLogLevel(level int) {
	log.SetLevel(log.Level(level))
}

func silenceLogging(silent bool) {
	log.Silence(silent)
}
