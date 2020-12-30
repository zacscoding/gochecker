package gochecker

import "encoding/json"

type Status string

const (
	up      Status = "UP"
	down    Status = "DOWN"
	unknown Status = "Unknown"
)

// HealthStatus represents result of aggregated component's health status
type HealthStatus struct {
	Status     Status                     `json:"status"`
	Components map[string]ComponentStatus `json:"components"`
}

// IsUp returns a true if status is up, otherwise false
func (s *HealthStatus) IsUp() bool {
	return s.Status == up
}

// IsDown returns a true if status is down, otherwise false
func (s *HealthStatus) IsDown() bool {
	return s.Status == down
}

// ComponentStatus represents result of one component's health status such as database
type ComponentStatus struct {
	status  Status
	details map[string]interface{}
}

func (s ComponentStatus) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{}
	data["status"] = s.status
	data["details"] = s.details
	return json.Marshal(data)
}

// WithUp updates status to "up"
func (s *ComponentStatus) WithUp() *ComponentStatus {
	s.status = up
	return s
}

// WithDown updates status to "down"
func (s *ComponentStatus) WithDown() *ComponentStatus {
	s.status = down
	return s
}

// WithDetail adds a detail with given key value
func (s *ComponentStatus) WithDetail(key string, value interface{}) *ComponentStatus {
	s.details[key] = value
	return s
}

// IsUp returns a true if status is up, otherwise false
func (s *ComponentStatus) IsUp() bool {
	return s.status == up
}

// IsDown returns a true if status is down, otherwise false
func (s *ComponentStatus) IsDown() bool {
	return s.status == down
}

// NewComponentStatus returns a new ComponentStatus with unknown status and empty details
func NewComponentStatus() *ComponentStatus {
	return &ComponentStatus{
		status:  unknown,
		details: make(map[string]interface{}),
	}
}
