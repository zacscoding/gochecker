package gochecker

import (
	"context"
	"sync"
	"time"
)

// Indicator is an interface to check health status.
type Indicator interface {

	// Health checks current indicator's health status.
	Health(ctx context.Context) ComponentStatus
}

type HealthFn func(ctx context.Context) ComponentStatus

type indicator struct {
	f HealthFn
}

func (i *indicator) Health(ctx context.Context) ComponentStatus {
	return i.f(ctx)
}

// HealthCheckerOption sets a parameter for health checker
type HealthCheckerOption func(r *CompositeHealthChecker)

// CompositeHealthChecker will check health status given checkers and observers
type CompositeHealthChecker struct {
	// indicators
	checkers  map[string]Indicator
	observers map[string]Indicator
	mutex     sync.Mutex
	// checker
	background         bool
	backgroundInterval time.Duration
	backgroundCtx      context.Context
	backgroundCancel   context.CancelFunc
	// health status
	cacheResult    *HealthStatus
	cacheValidTime time.Time
	cacheTTL       time.Duration
}

// AddChecker adds a health checker with given name, this indicator affect aggregated health status
func (c *CompositeHealthChecker) AddChecker(name string, checker Indicator) {
	c.checkers[name] = checker
}

// AddCheckerFn adds a health checker for given name and health check function, this indicator affect aggregated health status
func (c *CompositeHealthChecker) AddCheckerFn(name string, f HealthFn) {
	c.checkers[name] = &indicator{f: f}
}

// AddObserver adds a health observer with given name, this indicator does not affect aggregated health status
func (c *CompositeHealthChecker) AddObserver(name string, checker Indicator) {
	c.observers[name] = checker
}

// AddObserverFn adds a health observer for given name and health check function, this indicator does not affect aggregated health status
func (c *CompositeHealthChecker) AddObserverFn(name string, f HealthFn) {
	c.observers[name] = &indicator{f: f}
}

// Health returns a HealthStatus of components managed from this health checker
func (c *CompositeHealthChecker) Health(ctx context.Context) *HealthStatus {
	cacheable := c.cacheTTL != 0
	if !cacheable {
		return c.doHealthCheck(ctx)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.cacheResult != nil && time.Now().Before(c.cacheValidTime) {
		return c.cacheResult
	}

	status := c.doHealthCheck(ctx)

	// update cache result
	c.cacheResult = status
	c.cacheValidTime = time.Now().Add(c.cacheTTL)
	return c.cacheResult
}

// Close terminates background health check loop if enabled
func (c *CompositeHealthChecker) Close() {
	if c.background {
		c.backgroundCancel()
	}
}

func (c *CompositeHealthChecker) loopChecker() {
	var (
		nextCheck time.Time
		ticker    = time.NewTicker(c.backgroundInterval)
	)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if time.Now().After(nextCheck) {
				// skip if not exist checkers or observers
				if len(c.checkers)+len(c.observers) == 0 {
					continue
				}
				c.mutex.Lock()
				c.cacheResult = c.doHealthCheck(context.Background())
				c.cacheValidTime = time.Now().Add(c.cacheTTL)
				c.mutex.Unlock()
				nextCheck = time.Now().Add(c.backgroundInterval)
			}
		case <-c.backgroundCtx.Done():
			return
		}
	}
}

func (c *CompositeHealthChecker) doHealthCheck(ctx context.Context) *HealthStatus {
	type state struct {
		name       string
		status     ComponentStatus
		observable bool
	}
	var (
		hs = HealthStatus{
			Status:     up,
			Components: make(map[string]ComponentStatus),
		}
		resultCh = make(chan state, len(c.checkers)+len(c.observers))
		wg       = sync.WaitGroup{}
	)

	// check indicators
	for name, indicator := range c.checkers {
		wg.Add(1)
		go func(name string, indicator Indicator) {
			defer wg.Done()
			resultCh <- state{name: name, status: indicator.Health(ctx), observable: false}
		}(name, indicator)
	}
	// check observers
	for name, indicator := range c.observers {
		wg.Add(1)
		go func(name string, indicator Indicator) {
			defer wg.Done()
			resultCh <- state{name: name, status: indicator.Health(ctx), observable: true}
		}(name, indicator)
	}

	wg.Wait()
	close(resultCh)

	// aggregate results
	for result := range resultCh {
		if !result.observable && !result.status.IsUp() {
			hs.Status = down
		}
		hs.Components[result.name] = result.status
	}

	return &hs
}

// WithBackground sets background and backgroundInterval.
// i.e will start a new goroutine with given interval to check health status
func WithBackground(interval time.Duration) HealthCheckerOption {
	return func(c *CompositeHealthChecker) {
		c.background = true
		c.backgroundInterval = interval
		c.cacheTTL = 2 * interval
	}
}

// WithCacheTTL sets cache ttl of health result
func WithCacheTTL(cacheTTL time.Duration) HealthCheckerOption {
	return func(c *CompositeHealthChecker) {
		c.cacheTTL = cacheTTL
	}
}

// NewHealthChecker returns a new CompositeHealthChecker with given options
func NewHealthChecker(opts ...HealthCheckerOption) *CompositeHealthChecker {
	checker := CompositeHealthChecker{
		checkers:           make(map[string]Indicator),
		observers:          make(map[string]Indicator),
		mutex:              sync.Mutex{},
		background:         false,
		backgroundInterval: 0,
		cacheResult:        nil,
		cacheTTL:           time.Duration(0),
	}

	// apply options
	for _, opt := range opts {
		opt(&checker)
	}
	if checker.background {
		checker.backgroundCtx, checker.backgroundCancel = context.WithCancel(context.Background())
		go checker.loopChecker()
	}
	return &checker
}
