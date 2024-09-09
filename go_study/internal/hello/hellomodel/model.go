package hellomodel

import (
	"context"
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

type HelloRepository interface {
	Save(ctx context.Context, entity *HelloData) (*HelloData, error)
	FindById(ctx context.Context, id any) (*HelloData, error)
	ListAll(ctx context.Context) []HelloData
}
