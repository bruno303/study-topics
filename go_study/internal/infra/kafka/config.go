package kafka

type Config struct {
	Host         string
	Topic        string
	GroupId      string
	QntConsumers int
	Handler      MessageHandler
}
