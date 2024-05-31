package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Application struct {
		Name       string `yaml:"name"`
		Version    string `yaml:"version"`
		Monitoring struct {
			TraceUrl string `yaml:"trace-url"`
		} `yaml:"monitoring"`
	} `yaml:"app"`
	Database struct {
		Host         string `yaml:"host"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
		Port         int    `yaml:"port"`
		DatabaseName string `yaml:"database-name"`
	} `yaml:"database"`
	Kafka   KafkaConfig `yaml:"kafka"`
	Workers struct {
		HelloProducer HelloProducerConfig `yaml:"hello-producer"`
	} `yaml:"workers"`
}

type KafkaConfig struct {
	Host      string              `yaml:"host"`
	Consumers KafkaConsumerConfig `yaml:"consumers"`
}

type KafkaConsumerConfig struct {
	GoStudy KafkaConsumerConfigDetail `yaml:"go-study"`
}

type KafkaConsumerConfigDetail struct {
	Host         string `yaml:"host"`
	Topic        string `yaml:"topic"`
	GroupId      string `yaml:"group-id"`
	QntConsumers int    `yaml:"qnt-consumers"`
	TraceEnabled bool   `yaml:"trace-enabled"`
}

type HelloProducerConfig struct {
	IntervalMillis int64  `yaml:"interval-millis"`
	Topic          string `yaml:"topic"`
}

func LoadConfig() *Config {
	cfg := &Config{}
	file, err := os.Open("config/config.yaml")
	if err != nil {
		panic(err)
	}
	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(&cfg); err != nil {
		panic(err)
	}
	return cfg
}
