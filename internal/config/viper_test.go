package config

import (
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Since we set defaults, even without config.yaml in the test CWD, it should load successfully.
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error loading default config, got %v", err)
	}

	if cfg.Server.Port != 8000 {
		t.Errorf("expected default server port 8000, got %d", cfg.Server.Port)
	}
	if cfg.Server.Mode != "debug" {
		t.Errorf("expected default server mode 'debug', got '%s'", cfg.Server.Mode)
	}
	if cfg.Server.Env != "develop" {
		t.Errorf("expected default server env 'develop', got '%s'", cfg.Server.Env)
	}
	if cfg.Postgres.Database != "business_chat" {
		t.Errorf("expected default mongo database 'business_chat', got '%s'", cfg.Postgres.Database)
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	// Set environment variables to override values
	t.Setenv("APP_SERVER_PORT", "9090")
	t.Setenv("APP_SERVER_MODE", "release")
	t.Setenv("APP_SERVER_ENV", "production")
	t.Setenv("APP_POSTGRES_DATABASE", "prod_db")
	t.Setenv("APP_POSTGRES_URI", "mongodb://admin:secret@mongo-host:27017")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected env overridden server port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Server.Mode != "release" {
		t.Errorf("expected env overridden server mode 'release', got '%s'", cfg.Server.Mode)
	}
	if cfg.Server.Env != "production" {
		t.Errorf("expected env overridden server env 'production', got '%s'", cfg.Server.Env)
	}
	if cfg.Postgres.Database != "prod_db" {
		t.Errorf("expected env overridden mongo database 'prod_db', got '%s'", cfg.Postgres.Database)
	}
	if cfg.Postgres.URI != "mongodb://admin:secret@mongo-host:27017" {
		t.Errorf("expected env overridden mongo URI, got '%s'", cfg.Postgres.URI)
	}
}

func TestLoadConfig_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
	}{
		{
			name: "invalid port (too large)",
			envs: map[string]string{"APP_SERVER_PORT": "70000"},
		},
		{
			name: "invalid port (too small)",
			envs: map[string]string{"APP_SERVER_PORT": "0"},
		},
		{
			name: "invalid mode",
			envs: map[string]string{"APP_SERVER_MODE": "invalid_mode"},
		},
		{
			name: "invalid env",
			envs: map[string]string{"APP_SERVER_ENV": "local"},
		},
		{
			name: "empty postgres database",
			envs: map[string]string{"APP_POSTGRES_DATABASE": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env and set test envs
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}

			cfg, err := LoadConfig()
			if err == nil {
				t.Errorf("expected validation error, got nil config: %+v", cfg)
			}
		})
	}
}

func TestConfig_GetMaskedMongoURI(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected string
	}{
		{
			name:     "with credentials",
			uri:      "mongodb://admin:secret_pass@localhost:27017/db?authSource=admin",
			expected: "mongodb://username_masked:password_masked@localhost:27017/db?authSource=admin",
		},
		{
			name:     "no credentials",
			uri:      "mongodb://localhost:27017",
			expected: "mongodb://localhost:27017",
		},
		{
			name:     "invalid URI",
			uri:      "invalid-uri::%%",
			expected: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Postgres: PostgresConfig{
					URI: tt.uri,
				},
			}
			masked := cfg.GetMaskedMongoURI()
			if masked != tt.expected {
				t.Errorf("expected masked URI '%s', got '%s'", tt.expected, masked)
			}
		})
	}
}
