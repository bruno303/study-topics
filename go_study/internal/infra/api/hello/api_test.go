package hello

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apphello "github.com/bruno303/study-topics/go-study/internal/application/hello"
	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/config"
	"go.uber.org/mock/gomock"
)

func TestSetupApi_WhenHelloApiDisabled_DoesNotRegisterRoutes(t *testing.T) {
	cfg := &config.Config{}
	cfg.Application.Hello.Api.Enabled = false

	mux := http.NewServeMux()
	SetupApi(cfg, mux, nil)

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	res := httptest.NewRecorder()

	mux.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, res.Code)
	}
}

func TestCreate_WhenServiceSucceeds_ReturnsJSONResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := apphello.NewMockHelloService(ctrl)

	expected := models.HelloData{Id: "id-1", Name: "Bruno id-1", Age: 18}
	svc.
		EXPECT().
		Hello(gomock.Any(), gomock.AssignableToTypeOf(apphello.HelloInput{})).
		DoAndReturn(func(_ any, input apphello.HelloInput) (models.HelloData, error) {
			if input.Id == "" {
				t.Fatal("expected generated id to be non-empty")
			}
			if input.Age < 0 || input.Age >= 100 {
				t.Fatalf("expected generated age in [0, 99], got %d", input.Age)
			}
			return expected, nil
		})

	h := create(svc)
	req := httptest.NewRequest(http.MethodPost, "/hello", nil)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	if got := res.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	var got models.HelloData
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Fatalf("unexpected response body: %v", err)
	}
	if got != expected {
		t.Fatalf("expected %+v, got %+v", expected, got)
	}
}

func TestCreate_WhenServiceReturnsError_ReturnsInternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := apphello.NewMockHelloService(ctrl)

	errExpected := errors.New("service failed")
	svc.
		EXPECT().
		Hello(gomock.Any(), gomock.AssignableToTypeOf(apphello.HelloInput{})).
		Return(models.HelloData{}, errExpected)

	h := create(svc)
	req := httptest.NewRequest(http.MethodPost, "/hello", nil)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, res.Code)
	}
	if res.Body.String() != errExpected.Error() {
		t.Fatalf("expected body %q, got %q", errExpected.Error(), res.Body.String())
	}
}

func TestCreate_WhenCalled_PropagatesRequestContextToService(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := apphello.NewMockHelloService(ctrl)

	type contextKey string
	ctxKey := contextKey("trace-id")
	reqCtx := context.WithValue(context.Background(), ctxKey, "trace-123")

	svc.
		EXPECT().
		Hello(gomock.Any(), gomock.AssignableToTypeOf(apphello.HelloInput{})).
		DoAndReturn(func(ctx context.Context, _ apphello.HelloInput) (models.HelloData, error) {
			if got := ctx.Value(ctxKey); got != "trace-123" {
				t.Fatalf("expected context value %q, got %v", "trace-123", got)
			}
			return models.HelloData{Id: "id-1", Name: "Bruno id-1", Age: 18}, nil
		})

	h := create(svc)
	req := httptest.NewRequest(http.MethodPost, "/hello", nil).WithContext(reqCtx)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
}

func TestListAll_WhenServiceSucceeds_ReturnsJSONResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := apphello.NewMockHelloService(ctrl)

	expected := []models.HelloData{{Id: "id-1", Name: "Bruno id-1", Age: 18}}
	svc.EXPECT().ListAll(gomock.Any()).Return(expected, nil)

	h := listAll(svc)
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	if got := res.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	var got []models.HelloData
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Fatalf("unexpected response body: %v", err)
	}
	if len(got) != len(expected) || got[0] != expected[0] {
		t.Fatalf("expected %+v, got %+v", expected, got)
	}
}

func TestListAll_WhenServiceReturnsError_ReturnsInternalServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := apphello.NewMockHelloService(ctrl)

	errExpected := errors.New("list failed")
	svc.EXPECT().ListAll(gomock.Any()).Return(nil, errExpected)

	h := listAll(svc)
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, res.Code)
	}
	if res.Body.String() != errExpected.Error() {
		t.Fatalf("expected body %q, got %q", errExpected.Error(), res.Body.String())
	}
}

func TestListAll_WhenCalled_PropagatesRequestContextToService(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := apphello.NewMockHelloService(ctrl)

	type contextKey string
	ctxKey := contextKey("correlation-id")
	reqCtx := context.WithValue(context.Background(), ctxKey, "corr-123")

	svc.
		EXPECT().
		ListAll(gomock.Any()).
		DoAndReturn(func(ctx context.Context) ([]models.HelloData, error) {
			if got := ctx.Value(ctxKey); got != "corr-123" {
				t.Fatalf("expected context value %q, got %v", "corr-123", got)
			}
			return []models.HelloData{}, nil
		})

	h := listAll(svc)
	req := httptest.NewRequest(http.MethodGet, "/hello", nil).WithContext(reqCtx)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
}
