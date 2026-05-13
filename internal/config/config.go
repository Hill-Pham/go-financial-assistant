package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port int

	DatabaseURL string

	GeminiAPIKey string

	EvolutionAPIURL   string
	EvolutionInstance string
	EvolutionAPIKey   string
	OwnerPhone        string

	AllowedNumbers map[string]struct{}

	AdminSecret string
}

func Load() (*Config, error) {

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env: %w", err)
	}

	cfg := &Config{}
	var errs []error

	portStr := getEnv("PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid PORT: %q - must be a number", portStr))
	}
	cfg.Port = port

	cfg.DatabaseURL = getEnv("DATABASE_URL", "")
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}

	cfg.GeminiAPIKey = getEnv("GEMINI_API_KEY", "")
	if cfg.GeminiAPIKey == "" {
		errs = append(errs, errors.New("GEMINI_API_KEY is required"))
	}

	cfg.EvolutionAPIURL = getEnv("EVOLUTION_API_URL", "http://evolution:8082")
	cfg.EvolutionInstance = getEnv("EVOLUTION_INSTANCE", "")
	if cfg.EvolutionInstance == "" {
		errs = append(errs, errors.New("EVOLUTION_INSTANCE is required"))
	}
	cfg.EvolutionAPIKey = getEnv("EVOLUTION_API_KEY", "")
	if cfg.EvolutionAPIKey == "" {
		errs = append(errs, errors.New("EVOLUTION_API_KEY is required"))
	}

	cfg.OwnerPhone = getEnv("OWNER_PHONE", "")
	if cfg.OwnerPhone == "" {
		errs = append(errs, errors.New("OWNER_PHONE is required"))
	}

	cfg.AllowedNumbers = parseAllowedNumbers(getEnv("ALLOWED_NUMBERS", ""))
	cfg.AdminSecret = getEnv("ADMIN_SECRET", "")

	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("invalid configuration:\n%w", err)
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func parseAllowedNumbers(raw string) map[string]struct{} {
	allowed := make(map[string]struct{})
	for _, n := range strings.Split(raw, ",") {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		if !strings.Contains(n, "@") {
			n = n + "@s.whatsapp.net"
		}
		allowed[n] = struct{}{}
	}
	return allowed
}
