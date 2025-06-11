package config

import "fmt"

type Database struct {
	Driver   string
	Host     string
	User     string
	Port     int
	Password string
	DBName   string
	SSLMode  string
}

func (d Database) GetDsn() string {
	if d.SSLMode == "" {
		d.SSLMode = "disable"
	}
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", d.Host, d.User, d.Password, d.DBName, d.Port, d.SSLMode)
}

type Redis struct {
	Host     string
	Port     int
	DB       int
	Password string
}

func (r Redis) GetAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
