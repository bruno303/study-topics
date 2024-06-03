package attr

type Attribute struct {
	Key   string
	Value string
}

func New(key string, value string) Attribute {
	return Attribute{Key: key, Value: value}
}
