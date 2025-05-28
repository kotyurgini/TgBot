package config

import (
	"flag"
	"os"

	cleanenv "github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	TgToken     string `yaml:"tg_token" env-required:"true"`
	TgAdmin     int64  `yaml:"tg_admin" env-required:"true"`
	DbDriver    string `yaml:"db_driver" env-default:"sqlite"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	OpenAiToken string `yaml:"openai_token" env-required:"true"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist:" + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config-path", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
