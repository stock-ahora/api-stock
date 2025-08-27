package config

import "os"

type Config struct {
	Addr        string
	DatabaseURL string
}

func Load() Config {
	return Config{
		//        Addr:        getEnv("ADDR", ":8080"),
		//        DatabaseURL: getEnv("DATABASE_URL", ""),
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
