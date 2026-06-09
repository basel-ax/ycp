package config

import (
	"os"
	"testing"
)

func TestLoadConfig_ValidFile(t *testing.T) {
	cfg, err := LoadConfig("../example.env")
	if err != nil {
		t.Fatalf("LoadConfig returned an error: %v", err)
	}

	if cfg.TotalLimit != 100 {
		t.Fatalf("Expected TotalLimit to be 100, got %d", cfg.TotalLimit)
	}
	if cfg.TimeLimit != 3600 {
		t.Fatalf("Expected TimeLimit to be 3600, got %d", cfg.TimeLimit)
	}
	expected := "what the fuck? help me! i am trapped inside a computer."
	if cfg.FinalComment != expected {
		t.Fatalf("Expected FinalComment to be %q, got %q", expected, cfg.FinalComment)
	}
	if cfg.APIConnection != "" {
		t.Fatalf("Expected APIConnection to be empty, got %q", cfg.APIConnection)
	}
	if cfg.RedisHost != "localhost" {
		t.Fatalf("Expected RedisHost to be 'localhost', got %q", cfg.RedisHost)
	}
	if cfg.RedisPort != "6379" {
		t.Fatalf("Expected RedisPort to be '6379', got %q", cfg.RedisPort)
	}
	if cfg.RedisDB != 0 {
		t.Fatalf("Expected RedisDB to be 0, got %d", cfg.RedisDB)
	}
	if cfg.RedisCount != 5 {
		t.Fatalf("Expected RedisCount to be 5, got %d", cfg.RedisCount)
	}
}

func TestGetEnvAsInt_DefaultValue(t *testing.T) {
	key := "TEST_CONFIG_DEFAULT_VALUE_XYZ_UNSET"
	_ = os.Unsetenv(key)

	got := getEnvAsInt(key, 42)
	if got != 42 {
		t.Fatalf("Expected default value 42, got %d", got)
	}
}

func TestGetEnvAsInt_InvalidValue(t *testing.T) {
	key := "TEST_CONFIG_INVALID_VALUE_XYZ"
	if err := os.Setenv(key, "not_an_int"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer os.Unsetenv(key)

	got := getEnvAsInt(key, 99)
	if got != 99 {
		t.Fatalf("Expected default value 99 on invalid input, got %d", got)
	}
}

func TestGetEnvAsInt_ValidValue(t *testing.T) {
	key := "TEST_CONFIG_VALID_VALUE_XYZ"
	if err := os.Setenv(key, "123"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer os.Unsetenv(key)

	got := getEnvAsInt(key, 0)
	if got != 123 {
		t.Fatalf("Expected value 123, got %d", got)
	}
}
