package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr         string
	DatabaseURL      string
	LogLevel         string
	WeatherProvider  string
	TelescopeDriver  string
	AlpacaBaseURL    string
	WeatherQuota     int
	CORSOrigins      []string
	PublicHost       string
	JWTSecret        string
	SeedDemo         bool
}

func Load() Config {
	c := Config{
		HTTPAddr:        env("HTTP_ADDR", ":8080"),
		DatabaseURL:     env("DATABASE_URL", "postgres://gotosky:gotosky@127.0.0.1:28351/gotosky?sslmode=disable"),
		LogLevel:        env("LOG_LEVEL", "info"),
		WeatherProvider: strings.ToLower(env("WEATHER_PROVIDER", "mock")),
		TelescopeDriver: strings.ToLower(env("TELESCOPE_DRIVER", "mock")),
		AlpacaBaseURL:   env("ALPACA_BASE_URL", "http://127.0.0.1:11111"),
		WeatherQuota:    envInt("WEATHER_DAILY_QUOTA", 2000),
		CORSOrigins:     splitCSV(env("CORS_ORIGINS", "http://localhost:28353,http://127.0.0.1:28353")),
		PublicHost:      env("PUBLIC_HOST", "localhost:28353"),
		JWTSecret:       env("JWT_SECRET", "gotosky-dev-secret-change-me"),
		SeedDemo:        env("SEED_DEMO", "1") != "0",
	}
	return c
}

func (c Config) WeatherIsMock() bool   { return c.WeatherProvider == "mock" }
func (c Config) TelescopeIsMock() bool { return c.TelescopeDriver != "alpaca" }

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
