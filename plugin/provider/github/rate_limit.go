package github

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	sync "github.com/sniperkit/xutil/plugin/concurrency/sync/debug"
	// "github.com/anacrolix/sync"
	// "github.com/abourget/llerrgroup"

	"github.com/sniperkit/xtask/plugin/rate"
	"github.com/sniperkit/xtask/util/runtime"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

/*
	Notes:
 	- for one token, Github's rate limit for authenticated requests is 5000 QPH = 83.3 QPM = 1.38 QPS = 720ms/query
 	- AbuseRateLimitError is not always present ?! to verify/confirm asap !
	Refs:
 	- https://github.com/pantheon-systems/baryon/blob/fa9913c62c058fd4db1a95b6aa138d964021ef2e/source/gh/gh.go#L294-L327
*/

type Rate struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

type rateLimitCategory uint8

// initTimer initialize client timer.
func (g *Github) initTimer(resp *github.Response) {
	defer funcTrack(time.Now())

	if resp != nil {
		g.counters.Increment("github.initTimer", 1)
		log.Println("initTimer=", (*resp).Reset.Time.Sub(time.Now())+time.Second*2)
		timer := time.NewTimer((*resp).Reset.Time.Sub(time.Now()) + time.Second*2)
		g.timer = timer
		return
	}
}

// isLimited check if the client is available.
func (g *Github) isLimited() bool {
	defer funcTrack(time.Now())

	rate, _, err := g.Client.RateLimits(context.Background())
	if err != nil {
		return true
	}

	g.counters.Increment("rate.limit.isLimited", 1)

	response := new(struct {
		Resource *github.RateLimits `json:"resource"`
	})
	response.Resource = rate

	if response.Resource != nil {
		g.mu.Lock()
		defer g.mu.Unlock()
		if response.Resource.Core != nil {
			g.rateLimits[coreCategory].Limit = response.Resource.Core.Limit
			g.rateLimits[coreCategory].Remaining = response.Resource.Core.Remaining
			g.rateLimits[coreCategory].Reset = response.Resource.Core.Reset.Time
			return false
		}
		if response.Resource.Search != nil {
			g.rateLimits[searchCategory].Remaining = response.Resource.Search.Remaining
			g.rateLimits[searchCategory].Limit = response.Resource.Search.Limit
			g.rateLimits[searchCategory].Reset = response.Resource.Search.Reset.Time
			return false
		}
	}

	return true
}

func (g *Github) rateLimiter() *rate.RateLimiter {
	defer funcTrack(time.Now())

	// g.mu.RLock()
	// defer g.mu.RUnlock()

	rl, ok := g.rateLimiters[g.ctoken]
	if !ok {
		limit := 50
		// limit := len(g.ctokens) * 200
		if g.ctoken == "" {
			limit = 5
		}
		// rl = rate.New(limit, time.Second)
		rl = rate.New(limit, time.Minute)
		g.rateLimiters[g.ctoken] = rl
	}
	return rl
}

func (g *Github) recoverAbuse(statusCode int, msg string) (bool, *time.Duration) {
	defer funcTrack(time.Now())

	if strings.Contains(msg, "abuse") && statusCode == 403 {
		return true, &defaultRetryDelay
	}
	return false, nil
}

func (g *Github) limitHandler(statusCode int, rate github.Rate, hdrs http.Header, err error) error {
	defer funcTrack(time.Now())

	if err != nil {
		g.counters.Increment("limit.handler.err", 1)
		var (
			e  *github.AbuseRateLimitError
			ok bool
		)

		if statusCode == 404 {
			return errors.New("g.limitHandler() url not exists...")
		}

		if statusCode == 401 {
			log.Println("unautorized access, debug", runtime.WhereAmI())
			// oldToken := g.ctoken
			// newToken := g.getNextToken(oldToken)
			// g.client = g.getClient(newToken)
			return nil
		}

		if v := hdrs["Retry-After"]; len(v) > 0 {
			// According to GitHub support, the "Retry-After" header value will be
			// an integer which represents the number of seconds that one should
			// wait before resuming making requests.
			retryAfterSeconds, _ := strconv.ParseInt(v[0], 10, 64) // Error handling is noop.
			retryAfter := time.Duration(retryAfterSeconds) * time.Second
			log.Println("error:", err.Error(), "Retry-After=", v, "retryAfterSeconds: ", retryAfterSeconds, "retryAfter=", retryAfter.Seconds(), "debug", runtime.WhereAmI())
			time.Sleep(retryAfter)
			return errors.New("g.limitHandler()... API abuse detected...")
		}

		// Get the underlying error, if this is a Wrapped error by the github.com/pkg/errors package.
		// If not, this will just return the error itself.
		underlyingErr := errors.Cause(err)

		if e, ok = err.(*github.AbuseRateLimitError); ok {
			log.Println("error:", err.Error(), "e:", e, "err.(*github.AbuseRateLimitError) have triggered an abuse detection mechanism.", underlyingErr, "debug", runtime.WhereAmI())
			time.Sleep(*e.RetryAfter)
			return errors.New("g.limitHandler()... API abuse detected...")
		}

		switch underlyingErr.(type) {
		case *github.RateLimitError:
			log.Println("g.limitHandler()...RateLimitError")
			// oldToken := g.ctoken
			// newToken := g.getNextToken(oldToken)
			// g.client = g.getClient(newToken)
			return nil

		case *github.AbuseRateLimitError:
			var e *github.AbuseRateLimitError
			log.Println("error:", err.Error(), "e:", e, "*github.AbuseRateLimitError.", underlyingErr, "*e.RetryAfter=", *e.RetryAfter, "debug", runtime.WhereAmI())
			time.Sleep(*e.RetryAfter)
			return errors.New("g.limitHandler()... API abuse detected...")

		default:
			if strings.Contains(err.Error(), "timeout") ||
				strings.Contains(err.Error(), "abuse detection") ||
				strings.Contains(err.Error(), "try again") {
				time.Sleep(time.Duration(random(150, 720)) * time.Millisecond)
				log.Println("error:", err.Error(), "underlyingErr.(type).default", underlyingErr, "debug", runtime.WhereAmI())
				// return errors.New("Temporary error detected...")
			}
			return err
		}

	} else {
		g.counters.Increment("limit.handler.none", 1)
		log.Println("statusCode:", statusCode, "current.token=", g.ctoken, "rate.remaining=", rate.Remaining) //, "counters", g.counters.Snapshot())
	}
	return nil
}

func ExceededRateLimit(client *github.Client) (bool, *time.Time) {
	defer funcTrack(time.Now())

	var buf bytes.Buffer
	sync.PrintLockTimes(&buf)
	log.Println("sync.PrintLockTimes=", buf.String())

	rate, _, err := client.RateLimits(context.Background())
	if err != nil {
		log.Errorln("Error checking rate limit (%d): %s\n", rate.Core.Remaining, err)
		return true, nil //, rate
	}

	log.Infof("GitHub API rate limit: %d.\n", rate.Core.Remaining, "resetAt=", rate.Core.Reset.Time)

	// Check for a margin sufficient to run both examples.
	if rate.Core.Remaining <= 10 {
		time.Sleep(time.Duration(random(1000, 2000)) * time.Millisecond)
		log.Warnf("Exceeded (or almost exceeded) GitHub API rate limit: %d. Try again later.\n", rate.Core.Remaining)
		return true, &rate.Core.Reset.Time
	}
	return false, nil
}

func CheckLimit(statusCode int, rate github.Rate, hdrs http.Header, err error) (error, bool) {
	defer funcTrack(time.Now())

	if err != nil {
		var (
			e  *github.AbuseRateLimitError
			ok bool
		)

		if rate.Remaining <= 10 {
			time.Sleep(time.Duration(random(150, 720)) * time.Millisecond)
			log.Println("API rate limit of 5000 low remaining, debug=", runtime.WhereAmI())
			return errorRateLimitLowLevel, true
		}

		if statusCode == 404 {
			return errors.New("url not existing..."), false
		}

		if statusCode == 401 {
			log.Println("unautorized access, debug=", runtime.WhereAmI())
			return errorRateLimitReached, true
		}

		if statusCode == 403 {
			time.Sleep(time.Duration(random(150, 720)) * time.Millisecond)
			log.Println("API rate limit of 5000 still exceeded, debug=", runtime.WhereAmI())
			return errorRateLimitReached, true
		}

		if v := hdrs["Retry-After"]; len(v) > 0 {
			// According to GitHub support, the "Retry-After" header value will be
			// an integer which represents the number of seconds that one should
			// wait before resuming making requests.
			retryAfterSeconds, _ := strconv.ParseInt(v[0], 10, 64) // Error handling is noop.
			retryAfter := time.Duration(retryAfterSeconds) * time.Second
			log.Println("error:", err.Error(), "Retry-After=", v, "retryAfterSeconds: ", retryAfterSeconds, "retryAfter=", retryAfter.Seconds(), "debug", runtime.WhereAmI())
			time.Sleep(retryAfter)
			return errors.New("API abuse detected..."), false
		}

		// Get the underlying error, if this is a Wrapped error by the github.com/pkg/errors package.
		// If not, this will just return the error itself.
		underlyingErr := errors.Cause(err)

		if e, ok = err.(*github.AbuseRateLimitError); ok {
			log.Println("error:", err.Error(), "e:", e, "err.(*github.AbuseRateLimitError) have triggered an abuse detection mechanism.", underlyingErr, "debug", runtime.WhereAmI())
			time.Sleep(*e.RetryAfter)
			return errors.New("API abuse detected..."), false
		}

		switch underlyingErr.(type) {
		case *github.RateLimitError:
			log.Println("RateLimitError, debug=", runtime.WhereAmI())
			return errorRateLimitReached, true

		case *github.AbuseRateLimitError:
			var e *github.AbuseRateLimitError
			log.Println("error:", err.Error(), "e:", e, "*github.AbuseRateLimitError.", underlyingErr, "*e.RetryAfter=", *e.RetryAfter, "debug", runtime.WhereAmI())
			time.Sleep(*e.RetryAfter)
			return errors.New("API abuse detected..."), false

		default:
			if strings.Contains(err.Error(), "timeout") ||
				strings.Contains(err.Error(), "abuse detection") ||
				strings.Contains(err.Error(), "try again") {
				time.Sleep(time.Duration(random(150, 720)) * time.Millisecond)
				log.Println("error:", err.Error(), "underlyingErr.(type).default", underlyingErr, "debug", runtime.WhereAmI())
			}
			return err, false
		}

	} else {
		log.Println("statusCode:", statusCode, "rate.remaining=", rate.Remaining)
	}
	return nil, false
}

func (g *Github) getTokenKey(token string) int {
	defer funcTrack(time.Now())

	// g.mu.Lock()
	g.counters.Increment("token.get.key", 1)
	log.Println("getTokenKey().g.ctokens: ", len(g.ctokens))
	for k, t := range g.ctokens {
		if t.Key == token {
			log.Println("getTokenKey().currentIdx: ", k)
			return k
		}
	}
	return 0
}

func (g *Github) getNextToken(token string) string {
	defer funcTrack(time.Now())

	log.Println("getNextToken().g.ctokens: ", len(g.ctokens))
	for k, t := range g.ctokens {
		if t.Key == token {
			log.Println("getNextToken().currentIdx: ", k)
			nextKey := k + 1
			if nextKey > len(g.ctokens) {
				key := g.ctokens[0].Key
				return key
			}
			key := g.ctokens[nextKey].Key
			return key
		}
	}
	key := g.ctokens[0].Key
	return key
}

func (g *Github) getNextTokenKey(token string) int {
	defer funcTrack(time.Now())

	log.Println("getNextTokenKey().g.ctokens: ", len(g.ctokens))
	for k, t := range g.ctokens {
		if t.Key == token {
			log.Println("currentIdx: ", k)
			nextKey := k + 1
			if nextKey > len(g.ctokens) {
				return 0
			}
			return nextKey
		}
	}
	return 0
}

func (g *Github) checkRateLimit(statusCode int, rate github.Rate) {
	defer funcTrack(time.Now())

	g.counters.Increment("rate.check.limit", 1)
	log.Println("statusCode:", statusCode, "current.token=", g.ctoken, "rate.remaining=", rate.Remaining)
	if statusCode == 403 && rate.Remaining <= 0 {
		sleep := time.Until(rate.Reset.Time) + (time.Second * 10)
		if rate.Limit == 0 && rate.Remaining == 0 && statusCode == 403 {
			sleep = defaultAbuseDelay
		}
		log.Println("checkRateLimit().rate", rate, "statusCode:", statusCode, " sleep", sleep, "duration", time.Duration(sleep), "abuse?", (rate.Limit == 0 && rate.Remaining == 0 && statusCode == 403))
		time.Sleep(sleep)
	}
}

// GetRateLimit helps keep track of the API rate limit.
func (g *Github) getRateLimit() (int, error) {
	defer funcTrack(time.Now())

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Client == nil {
		g.Client = getClient(g.ctoken)
	}

	g.counters.Increment("rate.get.limit", 1)
	limits, _, err := g.Client.RateLimits(context.Background())
	if err != nil {
		return 0, err
	}
	return limits.Core.Limit, nil
}

// ex: rateLimiter("your_username").Wait()
func rateLimiter(name string) *rate.RateLimiter {
	defer funcTrack(time.Now())

	rl, ok := rateLimiters[name]
	if !ok {
		limit := 60
		if name == "" {
			limit = 5
		}
		rl = rate.New(limit, time.Minute)
		rateLimiters[name] = rl
	}
	return rl
}

func checkRateLimit(statusCode int, rate github.Rate) {
	defer funcTrack(time.Now())

	if rate.Remaining <= 10 || statusCode == 403 {
		sleep := time.Until(rate.Reset.Time) + (time.Second * 10)
		log.Println("statusCode:", statusCode, "checkRateLimit().rate", rate, " sleep", sleep, "duration", time.Duration(sleep))
		time.Sleep(sleep)
	}
}
