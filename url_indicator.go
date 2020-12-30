package gochecker

import (
	"context"
	"net/http"
	"time"
)

// UrlIndicator checks health status given method and url via http client
type UrlIndicator struct {
	Name    string
	Method  string
	Url     string
	Headers http.Header
	timeout time.Duration
}

func (u *UrlIndicator) Health(ctx context.Context) ComponentStatus {
	var timeout time.Duration
	deadline, ok := ctx.Deadline()
	if ok {
		timeout = deadline.Sub(time.Now())
	} else {
		timeout = u.timeout
	}

	var (
		cli = http.Client{
			Timeout: timeout,
		}
		status = NewComponentStatus()
	)

	req, err := http.NewRequest(u.Method, u.Url, nil)
	if err != nil {
		return *status.WithDown().WithDetail("err", err.Error())
	}
	if len(u.Headers) != 0 {
		for key, values := range u.Headers {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	resp, err := cli.Do(req)
	if err != nil {
		return *status.WithDown().WithDetail("err", err.Error())
	}
	if resp.StatusCode == http.StatusOK {
		status.WithUp()
	} else {
		status.WithDown()
	}
	return *status.WithDetail("status", resp.StatusCode)
}

// NewUrlIndicator creates a new url indicator with given args
func NewUrlIndicator(name, method, url string, headers http.Header, timeout time.Duration) Indicator {
	return &UrlIndicator{
		Name:    name,
		Method:  method,
		Url:     url,
		Headers: headers,
		timeout: timeout,
	}
}
