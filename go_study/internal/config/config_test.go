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

func TestLoadConfig_WhenOutboxSenderMaxAttemptsIsMissing_Panics(t *testing.T) {
	t.Setenv("CONFIG_FILE", "test-outbox-missing.yaml")

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected panic for missing outbox sender max-attempts")
		}
	}()

	_ = LoadConfig()
}

func TestLoadConfig_WhenOutboxSenderValuesAreMissingExceptMaxAttempts_DefaultsAreApplied(t *testing.T) {
	t.Setenv("CONFIG_FILE", "test-outbox-missing.yaml")
	t.Setenv("WORKERS_OUTBOX_SENDER_MAX_ATTEMPTS", "5")

	cfg := LoadConfig()

	if cfg.Workers.OutboxSender.IntervalMillis != defaultOutboxSenderPollIntervalMillis {
		t.Fatalf("expected interval-millis default %d, got %d", defaultOutboxSenderPollIntervalMillis, cfg.Workers.OutboxSender.IntervalMillis)
	}
	if cfg.Workers.OutboxSender.BatchSize != defaultOutboxSenderBatchSize {
		t.Fatalf("expected batch-size default %d, got %d", defaultOutboxSenderBatchSize, cfg.Workers.OutboxSender.BatchSize)
	}
	if cfg.Workers.OutboxSender.MaxAttempts != 5 {
		t.Fatalf("expected max-attempts override 5, got %d", cfg.Workers.OutboxSender.MaxAttempts)
	}
	if cfg.Workers.OutboxSender.RetryIntervalMillis != defaultOutboxSenderRetryIntervalMillis {
		t.Fatalf("expected retry-interval-millis default %d, got %d", defaultOutboxSenderRetryIntervalMillis, cfg.Workers.OutboxSender.RetryIntervalMillis)
	}
}

func TestLoadConfig_WhenOutboxSenderMaxAttemptsIsZeroOrNegative_Panics(t *testing.T) {
	testCases := []struct {
		name   string
		envVal string
	}{
		{name: "zero", envVal: "0"},
		{name: "negative", envVal: "-1"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("CONFIG_FILE", "test.yaml")
			t.Setenv("WORKERS_OUTBOX_SENDER_MAX_ATTEMPTS", testCase.envVal)

			defer func() {
				recovered := recover()
				if recovered == nil {
					t.Fatalf("expected panic for outbox sender max-attempts %s", testCase.envVal)
				}
			}()

			_ = LoadConfig()
		})
	}
}
