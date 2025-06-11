package config

import "fmt"

type Redis struct {
	Host     string
	Port     int
	DB       int
	Password string
}

func (r Redis) GetAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
