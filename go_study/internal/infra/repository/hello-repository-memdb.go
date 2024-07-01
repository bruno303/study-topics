package repository

import (
	"github.com/bruno303/study-topics/go-study/internal/hello"
	"github.com/bruno303/study-topics/go-study/internal/infra/database"
)

func NewHelloRepository2() *database.MemDbRepository[hello.HelloData] {
	return database.NewMemDbRepository[hello.HelloData]()
}
