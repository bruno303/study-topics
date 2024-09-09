package hellofile

import (
	"context"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
)

type HelloFileService struct {
	fileName *string
}

func NewHelloFileService() HelloFileService {
	return HelloFileService{}
}

func (s *HelloFileService) WriteFile(ctx context.Context, name string) error {
	s.fileName = &name
	log.Log().Info(ctx, "Writing file %s", *s.fileName)
	return nil
}

func (s *HelloFileService) Rollback(ctx context.Context) error {
	log.Log().Info(ctx, "Deleting file %s", *s.fileName)
	return nil
}
