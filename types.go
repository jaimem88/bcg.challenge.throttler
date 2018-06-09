package throttler

import "time"

type tokensCache map[string]*requester // unsafe in memory token cache

// Service throttler config
type Service struct {
	environment string
	Throttler   *Throttler
}

// NewService returns a new server instance with throttler config
func NewService(environment string, n, m int64) *Service {
	return &Service{
		environment: environment,
		Throttler: &Throttler{
			N:     n,
			M:     m,
			cache: tokensCache{},
		},
	}

}

// Throttler configuration
type Throttler struct {
	N     int64 `json:"n,omitempty"` // N is the number of requests allowed per M
	M     int64 `json:"m,omitempty"` // M milliseconds
	cache tokensCache
}

type requester struct {
	token     string
	counter   int64
	startTime time.Time
	endTime   time.Time
}
