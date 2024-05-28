package config

import "os"

type Config struct {
	Port                 string
	IdentityProviderHost string
	DBUsername           string
	DBPassword           string
	DBHost               string
	DBPort               string
}

func NewConfig() *Config {
	return &Config{
		Port:                 os.Getenv("SERVICE_PORT"),
		IdentityProviderHost: os.Getenv("IDENTITY_PROVIDER_HOST"),
		DBUsername:           os.Getenv("MONGO_INITDB_ROOT_USERNAME"),
		DBPassword:           os.Getenv("MONGO_INITDB_ROOT_PASSWORD"),
		DBHost:               os.Getenv("DB_HOST"),
		DBPort:               os.Getenv("DB_PORT"),
	}
}
