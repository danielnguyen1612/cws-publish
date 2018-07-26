package tools

import (
	"strings"

	"github.com/minio/go-homedir"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func init() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match
}

// InitViperConfig configure viper with config file
func InitViperConfig(cfgFile string, configName string, log *zap.Logger) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal("Can't get home directory", zap.Error(err))
		}

		viper.AddConfigPath(home)
		viper.SetConfigName("." + configName)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug("Initialize config", zap.String("filename", viper.ConfigFileUsed()))
	}
}
