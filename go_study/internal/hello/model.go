package hello

import (
	"fmt"
)

type HelloData struct {
	Id   string
	Name string
	Age  int
}

func (d HelloData) Key() string {
	return d.Id
}

func (d HelloData) ToString() string {
	return fmt.Sprintf("HelloData[id=%s, name=%s, age=%d]", d.Id, d.Name, d.Age)
}
