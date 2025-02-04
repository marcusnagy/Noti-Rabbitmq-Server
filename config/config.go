package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

const (
	httpPort   = "PORTS.HTTP.Port"
	healthPort = "PORTS.HEALTH.Port"
)

type Conf interface {
	Validate() error
}

type Config struct {
	Ports Ports `mapstructure:"ports"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Validate() error {
	if err := c.Ports.Validate(); err != nil {
		return err
	}
	return nil
}

func LoadConfig() *Config {
	cfg := NewConfig()

	viper, err := viperSetup()
	if err != nil {
		log.Fatalf("Failed to setup viper: %v", err)
	}

	getValues(viper, cfg)

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Failed to validate config: %v", err)
	}

	return cfg
}

func viperSetup() (*viper.Viper, error) {
	root := viper.New()
	root.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	root.AutomaticEnv()

	root.SetDefault(httpPort, 8080)
	root.SetDefault(healthPort, 8081)

	return root, nil
}

func getValues(root *viper.Viper, cfg *Config) {
	cfg.Ports.HTTPPort = root.GetInt(httpPort)
	cfg.Ports.HealthPort = root.GetInt(healthPort)
}
