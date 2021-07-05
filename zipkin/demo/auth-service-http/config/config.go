package config

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	_viper "github.com/spf13/viper"
)

var Config = _viper.New()

func SetupConfig() {
	Config.SetDefault("port", 3000)
	Config.SetDefault("request_logging", true)
	Config.SetDefault("jwt_rsa_public_key", "")
	Config.SetDefault("jwt_rsa_private_key", "")

	Config.AutomaticEnv()

	Config.SetConfigName("auth-service-http")
	Config.SetConfigType("json")
	Config.AddConfigPath("/etc/mock")
	Config.AddConfigPath(".")

	err := Config.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	log.Printf("Using config file: %s\n", Config.ConfigFileUsed())
	log.Printf("Final config: %v\n", Config.AllSettings())

	Config.WatchConfig()

	Config.OnConfigChange(func(in fsnotify.Event) {
		log.Println("Config file changes applied. (Note: port changes only apply when service restarts)")
		log.Printf("New config: %v\n", Config.AllSettings())
	})
}
