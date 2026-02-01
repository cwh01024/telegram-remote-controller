package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test loading config
	cfg := Load()
	if cfg == nil {
		t.Fatal("Load() returned nil")
	}
}

func TestValidateWithToken(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")

	cfg := Load()
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() should pass with token set, got: %v", err)
	}
}

func TestValidateWithoutToken(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")

	cfg := Load()
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should fail without token")
	}
}
