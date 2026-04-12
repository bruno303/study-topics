package handlers

import (
	"context"
	"errors"
	"testing"

	apphello "github.com/bruno303/study-topics/go-study/internal/application/hello"
	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"go.uber.org/mock/gomock"
)

func TestHelloMessageHandler_Process_WhenMessageIsInvalidJSON_ReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := apphello.NewMockHelloService(ctrl)
	handler := NewHelloMessageHandler(svc)

	svc.EXPECT().Hello(gomock.Any(), gomock.Any()).Times(0)

	err := handler.Process(context.Background(), "{invalid-json}")
	if err == nil {
		t.Fatal("expected error for invalid json, got nil")
	}
}

func TestHelloMessageHandler_Process_WhenServiceReturnsError_PropagatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := apphello.NewMockHelloService(ctrl)
	handler := NewHelloMessageHandler(svc)

	errExpected := errors.New("service failed")
	svc.
		EXPECT().
		Hello(gomock.Any(), gomock.Eq(apphello.HelloInput{Id: "abc", Age: 33})).
		Return(models.HelloData{}, errExpected)

	err := handler.Process(context.Background(), `{"id":"abc","age":33}`)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errExpected) {
		t.Fatalf("expected error %v, got %v", errExpected, err)
	}
}

func TestHelloMessageHandler_Process_WhenMessageIsValid_CallsServiceAndReturnsNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := apphello.NewMockHelloService(ctrl)
	handler := NewHelloMessageHandler(svc)

	svc.
		EXPECT().
		Hello(gomock.Any(), gomock.Eq(apphello.HelloInput{Id: "abc", Age: 33})).
		Return(models.HelloData{Id: "abc", Name: "Bruno abc", Age: 33}, nil)

	err := handler.Process(context.Background(), `{"id":"abc","age":33}`)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestHelloMessageHandler_Process_WhenMessageIsValid_PropagatesContextToService(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := apphello.NewMockHelloService(ctrl)
	handler := NewHelloMessageHandler(svc)

	type contextKey string
	ctxKey := contextKey("trace-id")
	incomingCtx := context.WithValue(context.Background(), ctxKey, "kafka-trace-001")

	svc.
		EXPECT().
		Hello(gomock.Any(), gomock.Eq(apphello.HelloInput{Id: "abc", Age: 33})).
		DoAndReturn(func(ctx context.Context, _ apphello.HelloInput) (models.HelloData, error) {
			if got := ctx.Value(ctxKey); got != "kafka-trace-001" {
				t.Fatalf("expected context value %q, got %v", "kafka-trace-001", got)
			}
			return models.HelloData{Id: "abc", Name: "Bruno abc", Age: 33}, nil
		})

	err := handler.Process(incomingCtx, `{"id":"abc","age":33}`)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
