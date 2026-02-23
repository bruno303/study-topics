package usecase

import (
	"context"
	"planning-poker/internal/application/planningpoker/metric"
	"planning-poker/internal/domain"

	"github.com/bruno303/go-toolkit/pkg/log"
	"github.com/google/uuid"
)

type (
	CreateClientOutput struct {
		ClientID string
	}
	CreateClientUseCase interface {
		Execute(ctx context.Context) (CreateClientOutput, error)
	}
	createClientUseCase struct {
		hub    domain.Hub
		logger log.Logger
		metric metric.PlanningPokerMetric
	}
)

var _ UseCaseO[CreateClientOutput] = (*createClientUseCase)(nil)

func NewCreateClientUseCase(hub domain.Hub, metric metric.PlanningPokerMetric) createClientUseCase {
	return createClientUseCase{
		hub:    hub,
		logger: log.NewLogger("usecase.CreateClient"),
		metric: metric,
	}
}

func (uc createClientUseCase) Execute(ctx context.Context) (CreateClientOutput, error) {
	clientID := uuid.NewString()
	uc.logger.Info(ctx, "Client created with ID: %s", clientID)

	return CreateClientOutput{ClientID: clientID}, nil
}
