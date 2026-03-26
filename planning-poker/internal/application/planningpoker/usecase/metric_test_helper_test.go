package usecase

import (
	"context"
	appmetric "planning-poker/internal/application/planningpoker/metric"
	toolkitmetric "github.com/bruno303/go-toolkit/pkg/metric"
	"reflect"
	"sync"
	"unsafe"
)

type mockMeter struct {
	mu    sync.Mutex
	calls []meterCall
}

type meterCall struct {
	name  string
	value float64
}

func (m *mockMeter) AddCounter(_ context.Context, name, _, _ string, value float64, _ ...toolkitmetric.Attribute) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, meterCall{name: name, value: value})
	return nil
}

func (m *mockMeter) AddGauge(_ context.Context, _, _, _ string, _ float64, _ ...toolkitmetric.Attribute) error {
	return nil
}

func (m *mockMeter) getCalls() []meterCall {
	m.mu.Lock()
	defer m.mu.Unlock()

	return append([]meterCall{}, m.calls...)
}

func newTestPlanningPokerMetric() (appmetric.PlanningPokerMetric, *mockMeter) {
	testMeter := &mockMeter{}
	planningPokerMetric := appmetric.NewPlanningPokerMetric()

	metricValue := reflect.ValueOf(&planningPokerMetric).Elem()
	meterField := metricValue.FieldByName("meter")
	reflect.NewAt(meterField.Type(), unsafe.Pointer(meterField.UnsafeAddr())).Elem().Set(reflect.ValueOf(testMeter))

	return planningPokerMetric, testMeter
}

func countMetricCalls(calls []meterCall, name string) int {
	count := 0
	for _, call := range calls {
		if call.name == name {
			count++
		}
	}

	return count
}
