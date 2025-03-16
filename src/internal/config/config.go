package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Addr string `yaml:"address" env-required:"true"`
}

type Config struct {
	Env        string `yaml:"env" env:"ENV" env-required:"true"`
	HTTPServer `yaml:"http_server"`
}

func MustLoad() *Config {
	var configPath string
	configPath = os.Getenv("CONFIG_PATH")
	if configPath == "" {
		flags := flag.String("config", "config/local.yaml", "Path to default config file")
		flag.Parse()

		configPath = *flags

		if configPath == "" {
			log.Fatalf("Config path is not set")
		}

		if _, err := os.Stat(configPath); os.IsNotExist((err)) {
			log.Fatalf("config file does not exists: %s", configPath)
		}

	}
	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("can not read config file: %s", err.Error())
	}

	return &cfg

}
