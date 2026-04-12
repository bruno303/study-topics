package config

import "testing"

func TestLoadConfig_WhenDatabaseDriverEnvIsBlank_DefaultsToPgxPool(t *testing.T) {
	t.Setenv("CONFIG_FILE", "test.yaml")
	t.Setenv("DATABASE_DRIVER", "   ")

	cfg := LoadConfig()

	if cfg.Database.Driver != DatabaseDriverPGXPool {
		t.Fatalf("expected default driver %q, got %q", DatabaseDriverPGXPool, cfg.Database.Driver)
	}
}

func TestLoadConfig_WhenTraceEnabledIsNotOverridden_DefaultsToFalse(t *testing.T) {
	t.Setenv("CONFIG_FILE", "test.yaml")

	cfg := LoadConfig()

	if cfg.Application.Monitoring.TraceEnabled {
		t.Fatal("expected trace to be disabled by default")
	}
}

func TestLoadConfig_WhenTraceEnabledEnvIsSet_OverridesYaml(t *testing.T) {
	t.Setenv("CONFIG_FILE", "test.yaml")
	t.Setenv("APPLICATION_MONITORING_TRACE_ENABLED", "true")

	cfg := LoadConfig()

	if !cfg.Application.Monitoring.TraceEnabled {
		t.Fatal("expected trace to be enabled from env override")
	}
}

func TestLoadConfig_WhenDatabaseDriverHasWhitespaceAndCase_NormalizesToMemDb(t *testing.T) {
	t.Setenv("CONFIG_FILE", "test.yaml")
	t.Setenv("DATABASE_DRIVER", "  MeMdB  ")

	cfg := LoadConfig()

	if cfg.Database.Driver != DatabaseDriverMemDB {
		t.Fatalf("expected normalized driver %q, got %q", DatabaseDriverMemDB, cfg.Database.Driver)
	}
}

func TestLoadConfig_WhenDatabaseDriverIsUnsupported_Panics(t *testing.T) {
	t.Setenv("CONFIG_FILE", "test.yaml")
	t.Setenv("DATABASE_DRIVER", "sqlite")

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic for unsupported database driver")
		}
	}()

	_ = LoadConfig()
}
