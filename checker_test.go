package gochecker

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type upIndicator struct {
	count int32
}

func (i *upIndicator) Health(_ context.Context) ComponentStatus {
	atomic.AddInt32(&i.count, 1)
	return *NewComponentStatus().WithUp()
}

type downIndicator struct {
	count int32
}

func (i *downIndicator) Health(_ context.Context) ComponentStatus {
	atomic.AddInt32(&i.count, 1)
	return *NewComponentStatus().WithDown()
}

func TestHealthChecker_Health(t *testing.T) {
	cases := []struct {
		Name      string
		Checkers  []Indicator
		Observers []Indicator
		IsUp      bool
	}{
		{
			Name:     "All Up Indicators",
			Checkers: []Indicator{&upIndicator{}, &upIndicator{}},
			IsUp:     true,
		}, {
			Name:     "All Down Indicators",
			Checkers: []Indicator{&downIndicator{}, &downIndicator{}},
			IsUp:     false,
		}, {
			Name:     "Up and Down Indicators",
			Checkers: []Indicator{&upIndicator{}, &downIndicator{}},
			IsUp:     false,
		}, {
			Name:      "All Up Observers",
			Observers: []Indicator{&upIndicator{}, &upIndicator{}},
			IsUp:      true,
		}, {
			Name:      "All Down Observers",
			Observers: []Indicator{&downIndicator{}, &downIndicator{}},
			IsUp:      true,
		}, {
			Name:      "Up and Down Observers",
			Observers: []Indicator{&upIndicator{}, &downIndicator{}},
			IsUp:      true,
		}, {
			Name:      "All Up Indicators with Down observers",
			Checkers:  []Indicator{&upIndicator{}, &upIndicator{}},
			Observers: []Indicator{&downIndicator{}, &upIndicator{}},
			IsUp:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			checker := NewHealthChecker()
			for i, indicator := range tc.Checkers {
				checker.AddChecker(fmt.Sprintf("checker-%d", i), indicator)
			}
			for i, indicator := range tc.Observers {
				checker.AddObserver(fmt.Sprintf("observer-%d", i), indicator)
			}

			status := checker.Health(context.Background())

			assert.Equal(t, tc.IsUp, status.IsUp())
		})
	}
}

func TestHealthChecker_Health_CacheTTL(t *testing.T) {
	var (
		ttl       = time.Millisecond * 200
		checker   = NewHealthChecker(WithCacheTTL(ttl))
		indicator = &upIndicator{}
	)
	checker.AddChecker("checker1", indicator)

	// when
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			checker.Health(context.Background())
		}()
	}
	wg.Wait()
	time.Sleep(ttl)
	checker.Health(context.Background())

	// then
	assert.EqualValues(t, 2, indicator.count)
}

func TestHealthChecker_Health_Background(t *testing.T) {
	var (
		interval  = time.Millisecond * 10
		checker   = NewHealthChecker(WithBackground(interval))
		indicator = &upIndicator{}
	)
	checker.AddChecker("checker1", indicator)
	time.Sleep(10 * interval)
	assert.Greater(t, int(indicator.count), 1)
}
