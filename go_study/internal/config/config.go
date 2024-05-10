package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Host         string `yaml:"host"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
		Port         int    `yaml:"port"`
		DatabaseName string `yaml:"database-name"`
	} `yaml:"database"`
	Kafka KafkaConfig
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
