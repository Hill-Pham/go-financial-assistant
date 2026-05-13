package config

import (
	"testing"
)

func setEnv(t *testing.T, vars map[string]string) {
	t.Helper()
	for k, v := range vars {
		t.Setenv(k, v)
	}
}

func validEnv() map[string]string {
	return map[string]string{
		"PORT":                     "8080",
		"DATABASE_URL":             "postgres://ube:pass@localhost/db",
		"GEMINI_API_KEY":    "gemini-key",
		"EVOLUTION_API_URL": "http://evolution:8080",
		"EVOLUTION_INSTANCE":       "my-instance",
		"EVOLUTION_API_KEY":        "evo-key",
		"OWNER_PHONE":              "5511999999999",
	}
}

func TestLoad_Success(t *testing.T) {
	setEnv(t, validEnv())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("port expected 8080, got %d", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://ube:pass@localhost/db" {
		t.Errorf("DatabaseURL incorrect: %s", cfg.DatabaseURL)
	}
	if cfg.GeminiAPIKey != "gemini-key" {
		t.Errorf("GeminiAPIKey incorrect: %s", cfg.GeminiAPIKey)
	}
	if cfg.EvolutionAPIURL != "http://evolution:8080" {
		t.Errorf("EvolutionAPIURL incorrect: %s", cfg.EvolutionAPIURL)
	}
	if cfg.EvolutionInstance != "my-instance" {
		t.Errorf("EvolutionInstance incorrect: %s", cfg.EvolutionInstance)
	}
	if cfg.EvolutionAPIKey != "evo-key" {
		t.Errorf("EvolutionAPIKey incorrect: %s", cfg.EvolutionAPIKey)
	}
	if cfg.OwnerPhone != "5511999999999" {
		t.Errorf("OwnerPhone incorrect: %s", cfg.OwnerPhone)
	}
}

func TestLoad_DefaultPort(t *testing.T) {
	env := validEnv()
	delete(env, "PORT")
	setEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("port default expected 8080, got %d", cfg.Port)
	}
}

func TestLoad_DefaultEvolutionAPIURL(t *testing.T) {
	env := validEnv()
	delete(env, "EVOLUTION_API_URL")
	setEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if cfg.EvolutionAPIURL != "http://evolution:8080" {
		t.Errorf("EvolutionAPIURL default incorrect: %s", cfg.EvolutionAPIURL)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	env := validEnv()
	env["PORT"] = "not-a-number"
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for PORT invalid")
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	env := validEnv()
	delete(env, "DATABASE_URL")
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for DATABASE_URL obrigatĂ³ria")
	}
}

func TestLoad_MissingGeminiAPIKey(t *testing.T) {
	env := validEnv()
	delete(env, "GEMINI_API_KEY")
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for GEMINI_API_KEY obrigatĂ³ria")
	}
}

func TestLoad_MissingEvolutionInstance(t *testing.T) {
	env := validEnv()
	delete(env, "EVOLUTION_INSTANCE")
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for EVOLUTION_INSTANCE obrigatĂ³ria")
	}
}

func TestLoad_MissingEvolutionAPIKey(t *testing.T) {
	env := validEnv()
	delete(env, "EVOLUTION_API_KEY")
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for EVOLUTION_API_KEY obrigatĂ³ria")
	}
}

func TestLoad_MissingOwnerPhone(t *testing.T) {
	env := validEnv()
	delete(env, "OWNER_PHONE")
	setEnv(t, env)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for OWNER_PHONE obrigatĂ³ria")
	}
}

func TestLoad_MultipleErrors(t *testing.T) {
	setEnv(t, map[string]string{
		"PORT": "invalid",
	})

	_, err := Load()
	if err == nil {
		t.Fatal("expected mĂºltiplos erros")
	}
}

func TestParseAllowedNumbers_Single(t *testing.T) {
	result := parseAllowedNumbers("5511999999999")
	if _, ok := result["5511999999999@s.whatsapp.net"]; !ok {
		t.Error("number not encontrado no mapa com sufixo @s.whatsapp.net")
	}
	if len(result) != 1 {
		t.Errorf("expected 1 income entry, got %d", len(result))
	}
}

func TestParseAllowedNumbers_Multiple(t *testing.T) {
	result := parseAllowedNumbers("111, 222, 333")
	for _, n := range []string{"111@s.whatsapp.net", "222@s.whatsapp.net", "333@s.whatsapp.net"} {
		if _, ok := result[n]; !ok {
			t.Errorf("number '%s' not encontrado", n)
		}
	}
	if len(result) != 3 {
		t.Errorf("expected 3 entries, got %d", len(result))
	}
}

func TestParseAllowedNumbers_AlreadyHasSuffix(t *testing.T) {
	result := parseAllowedNumbers("5511999999999@s.whatsapp.net")
	if _, ok := result["5511999999999@s.whatsapp.net"]; !ok {
		t.Error("number com sufixo existente not deve be duplicado")
	}
	if len(result) != 1 {
		t.Errorf("expected 1 income entry, got %d", len(result))
	}
}

func TestParseAllowedNumbers_Empty(t *testing.T) {
	result := parseAllowedNumbers("")
	if len(result) != 0 {
		t.Errorf("expected mapa empty, got %d entries", len(result))
	}
}

func TestLoad_AllowedNumbers(t *testing.T) {
	env := validEnv()
	env["ALLOWED_NUMBERS"] = "5511111111111, 5522222222222"
	setEnv(t, env)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if len(cfg.AllowedNumbers) != 2 {
		t.Errorf("expected 2 numbers permitidos, got %d", len(cfg.AllowedNumbers))
	}
}


