package hello

import (
	"context"
	"fmt"
	"strconv"
)

type Repository interface {
	Save(ctx context.Context, entity *HelloData) (*HelloData, error)
	FindById(ctx context.Context, id any) (*HelloData, error)
	ListAll(ctx context.Context) []HelloData
	BeginTransactionWithContext(ctx context.Context) (context.Context, error)
	Rollback(ctx context.Context)
	Commit(ctx context.Context)
	RunWithTransaction(ctx context.Context, callback func(context.Context) (*HelloData, error)) (*HelloData, error)
}

type HelloService struct {
	repository Repository
}

func NewService(repo Repository) HelloService {
	return HelloService{repository: repo}
}

func (s HelloService) HelloV1(ctx context.Context, id int) string {
	// fmt.Println("Executing service")

	ctxWithTransaction, err := s.repository.BeginTransactionWithContext(ctx)
	if err != nil {
		panic(1)
	}
	defer s.repository.Rollback(ctxWithTransaction)

	s.repository.Save(ctxWithTransaction, &HelloData{Id: strconv.Itoa(id), Name: fmt.Sprintf("Bruno %v", id)})

	// go s.listAll(ctxWithTransaction)
	found, err := s.repository.FindById(ctxWithTransaction, strconv.Itoa(id))
	if err != nil {
		return err.Error()
	}
	fmt.Printf("Data: %v\n", found)
	s.repository.Commit(ctxWithTransaction)
	return found.Name
}

func (s HelloService) Hello(ctx context.Context, id string, age int) string {
	data, err := s.repository.RunWithTransaction(ctx, func(ctxTx context.Context) (*HelloData, error) {
		newHello := HelloData{Id: id, Name: fmt.Sprintf("Bruno %v", id), Age: age}
		helloAdded, err := s.repository.Save(ctxTx, &newHello)
		if err != nil {
			return nil, err
		}
		helloFound, err := s.repository.FindById(ctxTx, helloAdded.Id)
		if err != nil {
			return nil, err
		}
		return helloFound, nil
	})
	if err != nil {
		return err.Error()
	}
	return data.ToString()
}

// func (s HelloService) listAll(ctx context.Context) {
// 	list := s.repository.ListAll(ctx)
// 	fmt.Printf("Printing %v rows\n", len(list))
// 	for i := range list {
// 		fmt.Printf("Data: %v\n", list[i])
// 	}
// }
