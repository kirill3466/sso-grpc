package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env             string        `yaml:"env" env-default:"local"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"15s"`
	GRPC            GRPCConfig    `yaml:"grpc"`
	SSO             SSOConfig     `yaml:"sso"`
	WriteUserIDs    []int64       `yaml:"write_user_ids"`
	WriteEmails     []string      `yaml:"write_emails"`
	StaleAfter      time.Duration `yaml:"stale_after" env-default:"1s"`
	Simulate        bool          `yaml:"simulate" env-default:"true"`
	SimulateEvery   time.Duration `yaml:"simulate_every" env-default:"100ms"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type SSOConfig struct {
	Addr string `yaml:"addr" env-required:"true"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file doesnt exist: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config" + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
