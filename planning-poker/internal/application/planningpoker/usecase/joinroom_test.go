package usecase

import (
	"context"
	"errors"
	"planning-poker/internal/application/lock"
	"planning-poker/internal/application/planningpoker/metric"
	"planning-poker/internal/domain"
	"planning-poker/internal/domain/entity"
	"planning-poker/internal/infra/boundaries/hub/clientcollection"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestNewJoinRoomUseCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHub := domain.NewMockHub(ctrl)
	mockLockManager := lock.NewMockLockManager(ctrl)
	mockMetric := metric.NewPlanningPokerMetric()

	uc := NewJoinRoomUseCase(mockHub, mockLockManager, mockMetric)

	if uc.hub != mockHub {
		t.Error("hub not set correctly")
	}
	if uc.lockManager != mockLockManager {
		t.Error("lockManager not set correctly")
	}
}

func TestJoinRoomUseCase_Execute_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockHub := domain.NewMockHub(ctrl)
	mockLockManager := lock.NewMockLockManager(ctrl)
	testMetric, metricMeter := newTestPlanningPokerMetric()
	mockBus := domain.NewMockBus(ctrl)

	roomID := "room123"
	room := &entity.Room{
		ID:      roomID,
		Clients: clientcollection.New(),
	}

	mockLockManager.EXPECT().
		WithLock(gomock.Any(), roomID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	mockHub.EXPECT().LoadRoom(ctx, roomID).Return(room, nil)
	mockHub.EXPECT().AddClient(gomock.Any())
	mockHub.EXPECT().AddBus(gomock.Any(), gomock.Any(), gomock.Any())

	mockBus.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	mockHub.EXPECT().BroadcastToRoom(ctx, roomID, gomock.Any()).Return(nil)

	uc := NewJoinRoomUseCase(mockHub, mockLockManager, testMetric)
	cmd := JoinRoomCommand{
		RoomID:   roomID,
		SenderID: "sender123",
		Bus:      mockBus,
	}

	output, err := uc.Execute(ctx, cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if output == nil {
		t.Fatal("expected output to be non-nil")
	}
	if output.Client == nil {
		t.Fatal("expected client to be non-nil")
	}
	if output.Room != room {
		t.Errorf("expected room %v, got %v", room, output.Room)
	}

	calls := metricMeter.getCalls()
	if countMetricCalls(calls, metric.PlanningPokerActiveRoomsMetric) != 0 {
		t.Fatalf("expected no active room increments for existing room, got %d", countMetricCalls(calls, metric.PlanningPokerActiveRoomsMetric))
	}
}

func TestJoinRoomUseCase_Execute_AutoCreatesMissingRoom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockHub := domain.NewMockHub(ctrl)
	mockBus := domain.NewMockBus(ctrl)
	mockLockManager := lock.NewMockLockManager(ctrl)
	testMetric, metricMeter := newTestPlanningPokerMetric()

	roomID := "nonexistent"
	room := &entity.Room{
		ID:      roomID,
		Clients: clientcollection.New(),
	}

	mockLockManager.EXPECT().
		WithLock(gomock.Any(), roomID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	mockHub.EXPECT().LoadRoom(ctx, roomID).Return(nil, domain.ErrRoomNotFound)
	mockHub.EXPECT().NewRoomWithID(ctx, roomID).Return(room, nil)
	mockHub.EXPECT().AddClient(gomock.Any())
	mockHub.EXPECT().AddBus(gomock.Any(), gomock.Any(), gomock.Any())
	mockBus.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)
	mockHub.EXPECT().BroadcastToRoom(ctx, roomID, gomock.Any()).Return(nil)

	uc := NewJoinRoomUseCase(mockHub, mockLockManager, testMetric)
	cmd := JoinRoomCommand{
		RoomID:   roomID,
		SenderID: "sender123",
		Bus:      mockBus,
	}

	output, err := uc.Execute(ctx, cmd)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if output == nil {
		t.Fatal("expected output to be non-nil")
	}
	if output.Room != room {
		t.Fatalf("expected room %v, got %v", room, output.Room)
	}
	if output.Client == nil {
		t.Fatal("expected client to be non-nil")
	}
	if !output.Client.IsOwner {
		t.Fatal("expected first auto-created room client to be owner")
	}

	calls := metricMeter.getCalls()
	if countMetricCalls(calls, metric.PlanningPokerActiveRoomsMetric) != 1 {
		t.Fatalf("expected one active room increment, got %d", countMetricCalls(calls, metric.PlanningPokerActiveRoomsMetric))
	}
}

func TestJoinRoomUseCase_Execute_LoadRoomError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockHub := domain.NewMockHub(ctrl)
	mockLockManager := lock.NewMockLockManager(ctrl)
	testMetric, metricMeter := newTestPlanningPokerMetric()
	mockBus := domain.NewMockBus(ctrl)

	roomID := "room123"
	expectedError := errors.New("redis unavailable")

	mockLockManager.EXPECT().
		WithLock(gomock.Any(), roomID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	mockHub.EXPECT().LoadRoom(ctx, roomID).Return(nil, expectedError)

	uc := NewJoinRoomUseCase(mockHub, mockLockManager, testMetric)
	cmd := JoinRoomCommand{
		RoomID:   roomID,
		SenderID: "sender123",
		Bus:      mockBus,
	}

	output, err := uc.Execute(ctx, cmd)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if output != nil {
		t.Fatal("expected nil output on load error")
	}
	if !errors.Is(err, expectedError) {
		t.Fatalf("expected error to wrap %v, got %v", expectedError, err)
	}
	if err.Error() != "failed to load room room123: redis unavailable" {
		t.Fatalf("unexpected error message: %v", err)
	}

	calls := metricMeter.getCalls()
	if len(calls) != 0 {
		t.Fatalf("expected no metric changes on load error, got %d calls", len(calls))
	}
}

func TestJoinRoomUseCase_Execute_AutoCreateRoomError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockHub := domain.NewMockHub(ctrl)
	mockLockManager := lock.NewMockLockManager(ctrl)
	testMetric, metricMeter := newTestPlanningPokerMetric()
	mockBus := domain.NewMockBus(ctrl)

	roomID := "room123"
	expectedError := errors.New("create room failed")

	mockLockManager.EXPECT().
		WithLock(gomock.Any(), roomID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	mockHub.EXPECT().LoadRoom(ctx, roomID).Return(nil, domain.ErrRoomNotFound)
	mockHub.EXPECT().NewRoomWithID(ctx, roomID).Return(nil, expectedError)

	uc := NewJoinRoomUseCase(mockHub, mockLockManager, testMetric)
	cmd := JoinRoomCommand{
		RoomID:   roomID,
		SenderID: "sender123",
		Bus:      mockBus,
	}

	output, err := uc.Execute(ctx, cmd)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if output != nil {
		t.Fatal("expected nil output on auto-create error")
	}
	if !errors.Is(err, expectedError) {
		t.Fatalf("expected error to wrap %v, got %v", expectedError, err)
	}
	if err.Error() != "failed to auto-create room room123: create room failed" {
		t.Fatalf("unexpected error message: %v", err)
	}

	calls := metricMeter.getCalls()
	if len(calls) != 0 {
		t.Fatalf("expected no metric changes on auto-create error, got %d calls", len(calls))
	}
}

func TestJoinRoomUseCase_Execute_SendError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockHub := domain.NewMockHub(ctrl)
	mockLockManager := lock.NewMockLockManager(ctrl)
	testMetric, metricMeter := newTestPlanningPokerMetric()
	mockBus := domain.NewMockBus(ctrl)

	roomID := "room123"
	room := &entity.Room{
		ID:      roomID,
		Clients: clientcollection.New(),
	}

	expectedError := errors.New("send failed")

	mockLockManager.EXPECT().
		WithLock(gomock.Any(), roomID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	mockHub.EXPECT().LoadRoom(ctx, roomID).Return(room, nil)
	mockHub.EXPECT().AddClient(gomock.Any())
	mockHub.EXPECT().AddBus(gomock.Any(), gomock.Any(), gomock.Any())
	mockBus.EXPECT().Send(gomock.Any(), gomock.Any()).Return(expectedError)
	mockHub.EXPECT().RemoveClient(ctx, gomock.Any(), roomID).Return(nil)

	uc := NewJoinRoomUseCase(mockHub, mockLockManager, testMetric)
	cmd := JoinRoomCommand{
		RoomID:   roomID,
		SenderID: "sender123",
		Bus:      mockBus,
	}

	output, err := uc.Execute(ctx, cmd)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if output != nil {
		t.Error("expected output to be nil when send fails")
	}

	calls := metricMeter.getCalls()
	if countMetricCalls(calls, metric.PlanningPokerUsersTotalMetric) != 2 {
		t.Fatalf("expected users total increment and rollback decrement, got %d calls", countMetricCalls(calls, metric.PlanningPokerUsersTotalMetric))
	}
	if countMetricCalls(calls, metric.PlanningPokerActiveUsersMetric) != 2 {
		t.Fatalf("expected active users increment and rollback decrement, got %d calls", countMetricCalls(calls, metric.PlanningPokerActiveUsersMetric))
	}
}

func TestJoinRoomUseCase_Execute_SendErrorOnAutoCreatedRoom_RollsBackJoinInitialization(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockHub := domain.NewMockHub(ctrl)
	mockLockManager := lock.NewMockLockManager(ctrl)
	testMetric, metricMeter := newTestPlanningPokerMetric()
	mockBus := domain.NewMockBus(ctrl)

	roomID := "auto-created-room"
	room := &entity.Room{
		ID:      roomID,
		Clients: clientcollection.New(),
	}
	expectedError := errors.New("send failed")

	mockLockManager.EXPECT().
		WithLock(gomock.Any(), roomID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	mockHub.EXPECT().LoadRoom(ctx, roomID).Return(nil, domain.ErrRoomNotFound)
	mockHub.EXPECT().NewRoomWithID(ctx, roomID).Return(room, nil)
	mockHub.EXPECT().AddClient(gomock.Any())
	mockHub.EXPECT().AddBus(gomock.Any(), gomock.Any(), gomock.Any())
	mockBus.EXPECT().Send(gomock.Any(), gomock.Any()).Return(expectedError)
	mockHub.EXPECT().RemoveClient(ctx, gomock.Any(), roomID).Return(nil)
	mockHub.EXPECT().LoadRoom(ctx, roomID).Return(nil, domain.ErrRoomNotFound)

	uc := NewJoinRoomUseCase(mockHub, mockLockManager, testMetric)
	cmd := JoinRoomCommand{
		RoomID:   roomID,
		SenderID: "sender123",
		Bus:      mockBus,
	}

	output, err := uc.Execute(ctx, cmd)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if output != nil {
		t.Fatal("expected nil output when rollback occurs")
	}
	if !errors.Is(err, expectedError) {
		t.Fatalf("expected error to wrap %v, got %v", expectedError, err)
	}

	calls := metricMeter.getCalls()
	if countMetricCalls(calls, metric.PlanningPokerActiveRoomsMetric) != 2 {
		t.Fatalf("expected active room increment and rollback decrement, got %d calls", countMetricCalls(calls, metric.PlanningPokerActiveRoomsMetric))
	}
	if countMetricCalls(calls, metric.PlanningPokerUsersTotalMetric) != 2 {
		t.Fatalf("expected users total increment and rollback decrement, got %d calls", countMetricCalls(calls, metric.PlanningPokerUsersTotalMetric))
	}
	if countMetricCalls(calls, metric.PlanningPokerActiveUsersMetric) != 2 {
		t.Fatalf("expected active users increment and rollback decrement, got %d calls", countMetricCalls(calls, metric.PlanningPokerActiveUsersMetric))
	}
}

func TestJoinRoomUseCase_Execute_SendErrorOnAutoCreatedRoom_WhenRollbackRoomCheckFails_DoesNotDecrementActiveRooms(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockHub := domain.NewMockHub(ctrl)
	mockLockManager := lock.NewMockLockManager(ctrl)
	testMetric, metricMeter := newTestPlanningPokerMetric()
	mockBus := domain.NewMockBus(ctrl)

	roomID := "auto-created-room"
	room := &entity.Room{
		ID:      roomID,
		Clients: clientcollection.New(),
	}
	expectedError := errors.New("send failed")
	loadError := errors.New("load room failed")

	mockLockManager.EXPECT().
		WithLock(gomock.Any(), roomID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, key string, fn func(context.Context) (any, error)) (any, error) {
			return fn(ctx)
		})

	mockHub.EXPECT().LoadRoom(ctx, roomID).Return(nil, domain.ErrRoomNotFound)
	mockHub.EXPECT().NewRoomWithID(ctx, roomID).Return(room, nil)
	mockHub.EXPECT().AddClient(gomock.Any())
	mockHub.EXPECT().AddBus(gomock.Any(), gomock.Any(), gomock.Any())
	mockBus.EXPECT().Send(gomock.Any(), gomock.Any()).Return(expectedError)
	mockHub.EXPECT().RemoveClient(ctx, gomock.Any(), roomID).Return(nil)
	mockHub.EXPECT().LoadRoom(ctx, roomID).Return(nil, loadError)

	uc := NewJoinRoomUseCase(mockHub, mockLockManager, testMetric)
	cmd := JoinRoomCommand{
		RoomID:   roomID,
		SenderID: "sender123",
		Bus:      mockBus,
	}

	output, err := uc.Execute(ctx, cmd)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if output != nil {
		t.Fatal("expected nil output when rollback occurs")
	}
	if !errors.Is(err, expectedError) {
		t.Fatalf("expected error to wrap %v, got %v", expectedError, err)
	}

	calls := metricMeter.getCalls()
	if countMetricCalls(calls, metric.PlanningPokerActiveRoomsMetric) != 1 {
		t.Fatalf("expected only the auto-create active room increment, got %d calls", countMetricCalls(calls, metric.PlanningPokerActiveRoomsMetric))
	}
	if countMetricCalls(calls, metric.PlanningPokerUsersTotalMetric) != 2 {
		t.Fatalf("expected users total increment and rollback decrement, got %d calls", countMetricCalls(calls, metric.PlanningPokerUsersTotalMetric))
	}
	if countMetricCalls(calls, metric.PlanningPokerActiveUsersMetric) != 2 {
		t.Fatalf("expected active users increment and rollback decrement, got %d calls", countMetricCalls(calls, metric.PlanningPokerActiveUsersMetric))
	}
}

func TestJoinRoomUseCase_Execute_NilDependencies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetric := metric.NewPlanningPokerMetric()

	tests := []struct {
		name string
		hub  domain.Hub
		lock lock.LockManager
	}{
		{name: "nil hub", hub: nil, lock: lock.NewMockLockManager(ctrl)},
		{name: "nil lockmanager", hub: domain.NewMockHub(ctrl), lock: nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatal("expected panic, but none occurred")
				}
			}()

			_ = NewJoinRoomUseCase(tc.hub, tc.lock, mockMetric)
		})
	}
}
