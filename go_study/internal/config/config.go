package config

type Config struct {
	Database *DatabaseConfig
}

type DatabaseConfig struct {
	Host         string
	User         string
	Password     string
	Port         int
	DatabaseName string
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
	}
}
