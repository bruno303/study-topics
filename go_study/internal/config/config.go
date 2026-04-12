package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"

	cfgconfig "github.com/bruno303/study-topics/go-study/config"
	envconfig "github.com/sethvargo/go-envconfig"
	"gopkg.in/yaml.v3"
)

const (
	DatabaseDriverPGXPool = "pgxpool"
	DatabaseDriverMemDB   = "memdb"
)

type Config struct {
	Application struct {
		Name    string `env:"NAME" yaml:"name"`
		Version string `env:"VERSION" yaml:"version"`
		Port    int    `env:"PORT" yaml:"port"`
		Hello   struct {
			Api struct {
				Enabled bool `env:"ENABLED" yaml:"enabled"`
			} `env:", prefix=API_" yaml:"api"`
		} `env:", prefix=HELLO_" yaml:"hello"`
		Monitoring struct {
			TraceEnabled bool   `env:"TRACE_ENABLED" yaml:"trace-enabled"`
			TraceUrl     string `env:"TRACE_URL" yaml:"trace-url"`
		} `env:", prefix=MONITORING_" yaml:"monitoring"`
		Log struct {
			Level  string `env:"LOG_LEVEL" yaml:"level"`
			Format string `env:"LOG_FORMAT" yaml:"format"`
		} `yaml:"log"`
		Auth struct {
			Enabled   bool   `env:"ENABLED" yaml:"enabled"`
			SecretKey string `env:"SECRET_KEY" yaml:"secret-key"`
		} `env:", prefix=AUTH_" yaml:"auth"`
	} `env:", prefix=APPLICATION_" yaml:"app"`
	Database struct {
		Driver       string `env:"DATABASE_DRIVER" yaml:"driver"`
		Host         string `env:"DATABASE_HOST" yaml:"host"`
		User         string `env:"DATABASE_USER" yaml:"user"`
		Password     string `env:"DATABASE_PASSWORD" yaml:"password"`
		Port         int    `env:"DATABASE_PORT" yaml:"port"`
		DatabaseName string `env:"DATABASE_NAME" yaml:"database-name"`
	} `yaml:"database"`
	Kafka   KafkaConfig `yaml:"kafka"`
	Workers struct {
		HelloProducer HelloProducerConfig `yaml:"hello-producer"`
	} `yaml:"workers"`
}

type KafkaConfig struct {
	Host      string              `env:"KAFKA_HOST" yaml:"host"`
	Consumers KafkaConsumerConfig `env:", prefix=KAFKA_CONSUMER_" yaml:"consumers"`
}

type KafkaConsumerConfig struct {
	GoStudy KafkaConsumerConfigDetail `env:", prefix=GO_STUDY_" yaml:"go-study"`
}

type KafkaConsumerConfigDetail struct {
	Host               string        `env:"HOST" yaml:"host"`
	Topic              string        `env:"TOPIC" yaml:"topic"`
	GroupId            string        `env:"GROUP_ID" yaml:"group-id"`
	QntConsumers       int           `env:"QNT_CONSUMERS" yaml:"qnt-consumers"`
	TraceEnabled       bool          `env:"TRACE_ENABLED" yaml:"trace-enabled"`
	Enabled            bool          `env:"ENABLED" yaml:"enabled"`
	AutoCommit         bool          `env:"AUTO_COMMIT" yaml:"auto-commit"`
	AutoCommitInterval time.Duration `env:"AUTO_COMMIT_INTERVAL" yaml:"auto-commit-interval"`
	OffsetReset        string        `env:"OFFSET_RESET" yaml:"offset-reset"`
	AsyncCommit        bool          `env:"ASYNC_COMMIT	" yaml:"async-commit"`
}

type HelloProducerConfig struct {
	IntervalMillis int64  `env:"WORKERS_HELLO_PRODUCER_INTERVAL_MILLIS" yaml:"interval-millis"`
	Topic          string `env:"WORKERS_HELLO_PRODUCER_TOPIC" yaml:"topic"`
	Enabled        bool   `env:"WORKERS_HELLO_PRODUCER_ENABLED" yaml:"enabled"`
	MaxMessages    int    `env:"WORKERS_HELLO_PRODUCER_MAX_MESSAGES" yaml:"max-messages"`
}

func LoadConfig() *Config {
	cfg := &Config{}

	cfgFile := "config.yaml"
	if os.Getenv("CONFIG_FILE") != "" {
		cfgFile = os.Getenv("CONFIG_FILE")
	}

	file, err := cfgconfig.ConfigFS.Open(cfgFile)
	if err != nil {
		panic(err)
	}
	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(cfg); err != nil {
		panic(err)
	}

	log.Log().Debug(context.TODO(), "config with yaml: %+v", cfg)

	if err = envconfig.ProcessWith(
		context.Background(),
		&envconfig.Config{
			Target:           cfg,
			DefaultOverwrite: true,
		},
	); err != nil {
		panic(err)
	}

	log.Log().Debug(context.TODO(), "config with envs: %+v", cfg)

	cfg.Database.Driver = strings.ToLower(strings.TrimSpace(cfg.Database.Driver))
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = DatabaseDriverPGXPool
	}

	if cfg.Database.Driver != DatabaseDriverPGXPool && cfg.Database.Driver != DatabaseDriverMemDB {
		panic(fmt.Sprintf("unsupported database driver %q, supported values: %s, %s", cfg.Database.Driver, DatabaseDriverPGXPool, DatabaseDriverMemDB))
	}

	return cfg
}
