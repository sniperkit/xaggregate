package github

import (
	"context"
	"time"

	"github.com/cenkalti/backoff"
)

func retryRegistrationFunc(f func() error) error {
	return f()
	// return nil
	defer funcTrack(time.Now())

	return backoff.Retry(f, backoff.WithMaxRetries(backoff.WithContext(backoff.NewConstantBackOff(defaultRetryDelay), context.Background()), defaultRetryAttempt))
}

func (g *Github) retryRegistrationFunc(f func() error) error {
	return nil
	defer funcTrack(time.Now())

	g.mu.Lock()
	defer g.mu.Unlock()

	return backoff.Retry(f, backoff.WithMaxRetries(backoff.WithContext(backoff.NewConstantBackOff(defaultRetryDelay), context.Background()), defaultRetryAttempt))
}

func (g *Github) retryNotifyRegistrationFunc(f func() error) error {
	return nil
	defer funcTrack(time.Now())

	g.mu.Lock()
	defer g.mu.Unlock()

	return backoff.RetryNotify(f, backoff.WithMaxRetries(backoff.WithContext(backoff.NewConstantBackOff(defaultRetryDelay), context.Background()), defaultRetryAttempt), g.notifyAttempts)
}

/*
import (
	"net/http"
	"time"

	"github.com/jpillora/backoff"
)

type requestCanceler interface {
	CancelRequest(*http.Request)
}

// RetryTransport will issue each request up to the specified number
// of retries, if an error is returned by the underlying
// RoundTripper. Note that HTTP errors (where the error value is nil)
// are not subject to retries.
type RetryTransport struct {
	Transport http.RoundTripper
	Retries   int

	// Delay is the amount of time to wait after an error before
	// retrying.
	Delay time.Duration
}

func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	b := &backoff.Backoff{
		Min:    t.Delay,
		Jitter: true,
	}

	var err error
	for try := 0; try < t.Retries; try++ {
		req2 := CloneRequest(req)
		var resp *http.Response
		resp, err = transport.RoundTrip(req2)
		if err == nil {
			return resp, nil
		}
		time.Sleep(b.Duration())
	}
	return nil, err
}

func (t *RetryTransport) CancelRequest(req *http.Request) {
	v, ok := t.Transport.(requestCanceler)
	if ok && v != nil {
		v.CancelRequest(req)
	}
}
*/
