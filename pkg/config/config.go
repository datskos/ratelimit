package config

import (
	"github.com/spf13/viper"
)

type AppConfig struct {
	Port        int
	Reflection  bool // use for testing with grpc_cli
	DatabaseDir string
}

func NewConfig() AppConfig {
	viper.SetDefault("port", 8080)
	viper.SetDefault("reflection", false)
	viper.SetDefault("database_dir", "./db")

	viper.AutomaticEnv()
	return AppConfig{
		Port:        viper.GetInt("port"),
		Reflection:  viper.GetBool("reflection"),
		DatabaseDir: viper.GetString("database_dir"),
	}
}
