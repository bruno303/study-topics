package config

type Config struct {
	Database *DatabaseConfig
	Kafka    *KafkaConfig
}

type DatabaseConfig struct {
	Host         string
	User         string
	Password     string
	Port         int
	DatabaseName string
}

type KafkaConfig struct {
	Host      string
	Consumers *KafkaConsumerConfig
}

type KafkaConsumerConfig struct {
	GoStudy *KafkaConsumerConfigDetail
}

type KafkaConsumerConfigDetail struct {
	Topic        string
	GroupId      string
	QntConsumers int
}

func LoadConfig() *Config {
	return &Config{
		Database: &DatabaseConfig{
			Host:         "localhost",
			User:         "postgres",
			Password:     "postgres",
			Port:         5432,
			DatabaseName: "hello",
		},
		Kafka: &KafkaConfig{
			Host: "localhost:9092",
			Consumers: &KafkaConsumerConfig{
				GoStudy: &KafkaConsumerConfigDetail{
					Topic:        "go-study.hello",
					GroupId:      "my-group-1",
					QntConsumers: 3,
				},
			},
		},
	}
}
