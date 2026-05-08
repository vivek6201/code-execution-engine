package config

import "os"

type Config struct {
	DbUrl     string
	RedisUrl  string
	JWTSecret string
}

func Load() *Config {
	return &Config{
		DbUrl:     os.Getenv("DB_URL"),
		RedisUrl:  os.Getenv("REDIS_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),
	}
}
