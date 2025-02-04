package config

import (
	"errors"
)

type Ports struct {
	HTTPPort   int `mapstructure:"http_port"`
	HealthPort int `mapstructure:"health_port"`
}

func (p *Ports) Validate() error {
	if p.HTTPPort == 0 {
		return errors.New("HTTPPort is required")
	}
	if p.HealthPort == 0 {
		return errors.New("HealthPort is required")
	}
	return nil
}
