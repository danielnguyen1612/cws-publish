package tools

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// Execute run a command
func Execute(rootCmd *cobra.Command, log *zap.Logger) {
	RecoverLog(log, func() {
		if err := rootCmd.Execute(); err != nil {
			log.Fatal("Couldn't run", zap.Error(err))
		}
	})
}

// PreExecuteConfiguration read a yaml config file
func PreExecuteConfiguration(rootCmd *cobra.Command, configName string, log *zap.Logger) {
	var cfgFile string
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/.%s.yaml)", configName))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	cobra.OnInitialize(func() {
		InitViperConfig(cfgFile, configName, log)
	})
}
