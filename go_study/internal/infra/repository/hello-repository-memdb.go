package repository

import (
	"main/internal/hello"
	"main/internal/infra/database"
)

func NewHelloRepository2() *database.MemDbRepository[hello.HelloData] {
	return database.NewMemDbRepository[hello.HelloData]()
}
