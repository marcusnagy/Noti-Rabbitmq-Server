package config

import "errors"

type RabbitMQConfig struct {
	URL string `mapstructure:"RABBITMQ_URL"`
}

func (c *RabbitMQConfig) Validate() error {
	if c.URL == "" {
		return errors.New("RABBITMQ_URL is required")
	}
	return nil
}
