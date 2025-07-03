package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "server-updater",
	Short: "Manage Minecraft server software and plugin dependencies",
	Long:  `server-updater is a CLI utility that downloads and manages Minecraft server jars and plugins as defined in a simple text manifest.`,
}

func ExecuteMain() {
	cobra.CheckErr(rootCmd.Execute())
}
